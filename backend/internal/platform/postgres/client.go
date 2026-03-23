package postgres

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrNoRows = errors.New("postgres: no rows")

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type Client struct {
	cfg Config
}

type Conn struct {
	conn net.Conn
	cfg  Config
}

type Tx struct {
	conn *Conn
}

type Row []string

type Result struct {
	Rows []Row
}

func ParseConfig(databaseURL string) (Config, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return Config{}, fmt.Errorf("parse postgres url: %w", err)
	}

	password, _ := u.User.Password()
	cfg := Config{
		Host:     u.Hostname(),
		Port:     u.Port(),
		User:     u.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(u.Path, "/"),
		SSLMode:  u.Query().Get("sslmode"),
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.SSLMode != "disable" {
		return Config{}, fmt.Errorf("unsupported sslmode %q: only disable is supported", cfg.SSLMode)
	}
	if cfg.Host == "" || cfg.User == "" || cfg.Database == "" {
		return Config{}, errors.New("postgres url must include host, user, and database")
	}

	return cfg, nil
}

func NewClient(databaseURL string) (*Client, error) {
	cfg, err := ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	return &Client{cfg: cfg}, nil
}

func (c *Client) Exec(ctx context.Context, query string) error {
	conn, err := c.Connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Query(ctx, query)
	return err
}

func (c *Client) Query(ctx context.Context, query string) ([]Row, error) {
	conn, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	result, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return result.Rows, nil
}

func (c *Client) QueryRow(ctx context.Context, query string) (Row, error) {
	rows, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNoRows
	}
	return rows[0], nil
}

func (c *Client) Begin(ctx context.Context) (*Tx, error) {
	conn, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Query(ctx, "BEGIN"); err != nil {
		conn.Close()
		return nil, err
	}
	return &Tx{conn: conn}, nil
}

func (c *Client) Connect(ctx context.Context) (*Conn, error) {
	dialer := &net.Dialer{}
	netConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(c.cfg.Host, c.cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("dial postgres: %w", err)
	}
	conn := &Conn{conn: netConn, cfg: c.cfg}
	if err := conn.startup(ctx); err != nil {
		_ = netConn.Close()
		return nil, err
	}
	return conn, nil
}

func (tx *Tx) Exec(ctx context.Context, query string) error {
	_, err := tx.conn.Query(ctx, query)
	return err
}

func (tx *Tx) Query(ctx context.Context, query string) ([]Row, error) {
	result, err := tx.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return result.Rows, nil
}

func (tx *Tx) QueryRow(ctx context.Context, query string) (Row, error) {
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNoRows
	}
	return rows[0], nil
}

func (tx *Tx) Commit(ctx context.Context) error {
	defer tx.conn.Close()
	_, err := tx.conn.Query(ctx, "COMMIT")
	return err
}

func (tx *Tx) Rollback(ctx context.Context) error {
	defer tx.conn.Close()
	_, err := tx.conn.Query(ctx, "ROLLBACK")
	return err
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Query(ctx context.Context, query string) (Result, error) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetDeadline(deadline)
		defer c.conn.SetDeadline(time.Time{})
	}

	payload := append([]byte(query), 0)
	if err := writeMessage(c.conn, 'Q', payload); err != nil {
		return Result{}, fmt.Errorf("send query: %w", err)
	}

	var rows []Row
	for {
		typ, msg, err := readMessage(c.conn)
		if err != nil {
			return Result{}, fmt.Errorf("read query response: %w", err)
		}
		switch typ {
		case 'T':
			continue
		case 'D':
			row, err := parseDataRow(msg)
			if err != nil {
				return Result{}, err
			}
			rows = append(rows, row)
		case 'C', 'I', 'n':
			continue
		case 'E':
			return Result{}, parseErrorResponse(msg)
		case 'Z':
			return Result{Rows: rows}, nil
		default:
			continue
		}
	}
}

func (c *Conn) startup(ctx context.Context) error {
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetDeadline(deadline)
		defer c.conn.SetDeadline(time.Time{})
	}

	payload := make([]byte, 0, 128)
	payload = appendInt32(payload, 196608)
	payload = appendCString(payload, "user")
	payload = appendCString(payload, c.cfg.User)
	payload = appendCString(payload, "database")
	payload = appendCString(payload, c.cfg.Database)
	payload = appendCString(payload, "client_encoding")
	payload = appendCString(payload, "UTF8")
	payload = append(payload, 0)

	length := int32(len(payload) + 4)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(length))
	if _, err := c.conn.Write(buf); err != nil {
		return fmt.Errorf("write startup length: %w", err)
	}
	if _, err := c.conn.Write(payload); err != nil {
		return fmt.Errorf("write startup payload: %w", err)
	}

	for {
		typ, msg, err := readMessage(c.conn)
		if err != nil {
			return fmt.Errorf("read startup response: %w", err)
		}
		switch typ {
		case 'R':
			if err := c.handleAuth(msg); err != nil {
				return err
			}
		case 'S', 'K', 'N':
			continue
		case 'E':
			return parseErrorResponse(msg)
		case 'Z':
			return nil
		default:
			continue
		}
	}
}

func (c *Conn) handleAuth(msg []byte) error {
	if len(msg) < 4 {
		return errors.New("invalid auth message")
	}
	code := int(binary.BigEndian.Uint32(msg[:4]))
	switch code {
	case 0:
		return nil
	case 3:
		return c.sendPassword(c.cfg.Password)
	case 5:
		if len(msg) < 8 {
			return errors.New("invalid md5 auth payload")
		}
		salt := msg[4:8]
		return c.sendPassword(md5Password(c.cfg.User, c.cfg.Password, salt))
	case 10:
		return c.handleSASL(msg[4:])
	case 11, 12:
		return nil
	default:
		return fmt.Errorf("unsupported auth method: %d", code)
	}
}

func (c *Conn) sendPassword(password string) error {
	payload := append([]byte(password), 0)
	if err := writeMessage(c.conn, 'p', payload); err != nil {
		return fmt.Errorf("send password: %w", err)
	}
	return nil
}

func (c *Conn) handleSASL(msg []byte) error {
	mechanisms := bytesToStrings(msg)
	found := false
	for _, mechanism := range mechanisms {
		if mechanism == "SCRAM-SHA-256" {
			found = true
			break
		}
	}
	if !found {
		return errors.New("postgres does not offer SCRAM-SHA-256")
	}

	nonceRaw := make([]byte, 18)
	if _, err := rand.Read(nonceRaw); err != nil {
		return fmt.Errorf("generate scram nonce: %w", err)
	}
	nonce := base64.StdEncoding.EncodeToString(nonceRaw)
	clientFirstBare := "n=" + scramEscape(c.cfg.User) + ",r=" + nonce
	clientFirst := "n,," + clientFirstBare

	payload := make([]byte, 0, len(clientFirst)+64)
	payload = appendCString(payload, "SCRAM-SHA-256")
	payload = appendInt32(payload, int32(len(clientFirst)))
	payload = append(payload, []byte(clientFirst)...)
	if err := writeMessage(c.conn, 'p', payload); err != nil {
		return fmt.Errorf("send scram init: %w", err)
	}

	for {
		typ, resp, err := readMessage(c.conn)
		if err != nil {
			return fmt.Errorf("read scram response: %w", err)
		}
		if typ == 'E' {
			return parseErrorResponse(resp)
		}
		if typ != 'R' || len(resp) < 4 {
			continue
		}
		code := int(binary.BigEndian.Uint32(resp[:4]))
		switch code {
		case 11:
			serverFirst := string(resp[4:])
			attrs := parseSCRAMAttributes(serverFirst)
			salt, err := base64.StdEncoding.DecodeString(attrs["s"])
			if err != nil {
				return fmt.Errorf("decode scram salt: %w", err)
			}
			iterations, err := strconv.Atoi(attrs["i"])
			if err != nil {
				return fmt.Errorf("parse scram iteration count: %w", err)
			}
			combinedNonce := attrs["r"]
			clientFinalWithoutProof := "c=biws,r=" + combinedNonce
			authMessage := clientFirstBare + "," + serverFirst + "," + clientFinalWithoutProof
			saltedPassword := scramSaltedPassword(c.cfg.Password, salt, iterations)
			clientKey := hmacSHA256(saltedPassword, []byte("Client Key"))
			storedKey := sha256.Sum256(clientKey)
			clientSignature := hmacSHA256(storedKey[:], []byte(authMessage))
			clientProof := xorBytes(clientKey, clientSignature)
			clientFinal := clientFinalWithoutProof + ",p=" + base64.StdEncoding.EncodeToString(clientProof)
			if err := writeMessage(c.conn, 'p', []byte(clientFinal)); err != nil {
				return fmt.Errorf("send scram proof: %w", err)
			}
		case 12:
			return nil
		case 0:
			return nil
		default:
			return fmt.Errorf("unexpected scram auth code: %d", code)
		}
	}
}

func writeMessage(w io.Writer, typ byte, payload []byte) error {
	buf := make([]byte, 5)
	buf[0] = typ
	binary.BigEndian.PutUint32(buf[1:], uint32(len(payload)+4))
	if _, err := w.Write(buf); err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	_, err := w.Write(payload)
	return err
}

func readMessage(r io.Reader) (byte, []byte, error) {
	head := make([]byte, 5)
	if _, err := io.ReadFull(r, head); err != nil {
		return 0, nil, err
	}
	typ := head[0]
	length := int(binary.BigEndian.Uint32(head[1:5]))
	if length < 4 {
		return 0, nil, errors.New("invalid message length")
	}
	payload := make([]byte, length-4)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	return typ, payload, nil
}

func parseDataRow(msg []byte) (Row, error) {
	if len(msg) < 2 {
		return nil, errors.New("invalid data row")
	}
	count := int(binary.BigEndian.Uint16(msg[:2]))
	msg = msg[2:]
	row := make(Row, 0, count)
	for i := 0; i < count; i++ {
		if len(msg) < 4 {
			return nil, errors.New("invalid data row column length")
		}
		length := int(int32(binary.BigEndian.Uint32(msg[:4])))
		msg = msg[4:]
		if length == -1 {
			row = append(row, "")
			continue
		}
		if len(msg) < length {
			return nil, errors.New("invalid data row column payload")
		}
		row = append(row, string(msg[:length]))
		msg = msg[length:]
	}
	return row, nil
}

func parseErrorResponse(msg []byte) error {
	parts := strings.Split(string(msg), "\x00")
	for _, part := range parts {
		if len(part) > 1 && part[0] == 'M' {
			return errors.New(part[1:])
		}
	}
	return errors.New("postgres returned an error")
}

func appendInt32(dst []byte, value int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return append(dst, buf...)
}

func appendCString(dst []byte, value string) []byte {
	dst = append(dst, []byte(value)...)
	return append(dst, 0)
}

func bytesToStrings(payload []byte) []string {
	parts := strings.Split(string(payload), "\x00")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func md5Password(user, password string, salt []byte) string {
	inner := md5.Sum([]byte(password + user))
	innerHex := hex.EncodeToString(inner[:])
	outerInput := append([]byte(innerHex), salt...)
	outer := md5.Sum(outerInput)
	return "md5" + hex.EncodeToString(outer[:])
}

func scramEscape(value string) string {
	value = strings.ReplaceAll(value, "=", "=3D")
	value = strings.ReplaceAll(value, ",", "=2C")
	return value
}

func parseSCRAMAttributes(input string) map[string]string {
	attrs := make(map[string]string)
	for _, part := range strings.Split(input, ",") {
		chunks := strings.SplitN(part, "=", 2)
		if len(chunks) == 2 {
			attrs[chunks[0]] = chunks[1]
		}
	}
	return attrs
}

func scramSaltedPassword(password string, salt []byte, iterations int) []byte {
	ui := hmacSHA256([]byte(password), append(salt, 0, 0, 0, 1))
	result := append([]byte(nil), ui...)
	for i := 1; i < iterations; i++ {
		ui = hmacSHA256([]byte(password), ui)
		for j := range result {
			result[j] ^= ui[j]
		}
	}
	return result
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func xorBytes(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}
