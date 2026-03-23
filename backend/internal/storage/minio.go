package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"infographic-generator/backend/internal/config"
	"infographic-generator/backend/internal/utils"
)

const defaultMinIORegion = "us-east-1"

type MinIOStorage struct {
	httpClient           *http.Client
	endpoint             string
	bucket               string
	accessKey            string
	secretKey            string
	useSSL               bool
	multipartThresholdMB int64
	multipartPartSizeMB  uint64
}

type initiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	UploadID string   `xml:"UploadId"`
}

type completeMultipartUpload struct {
	XMLName xml.Name                `xml:"CompleteMultipartUpload"`
	Parts   []completeMultipartPart `xml:"Part"`
}

type completeMultipartPart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

func NewMinIOStorage(ctx context.Context, cfg config.Config) (*MinIOStorage, error) {
	storage := &MinIOStorage{
		httpClient:           &http.Client{Timeout: 60 * time.Second},
		endpoint:             cfg.MinIOEndpoint,
		bucket:               cfg.MinIOBucket,
		accessKey:            cfg.MinIOAccessKey,
		secretKey:            cfg.MinIOSecretKey,
		useSSL:               cfg.MinIOUseSSL,
		multipartThresholdMB: cfg.MultipartThresholdMB,
		multipartPartSizeMB:  cfg.MultipartPartSizeMB,
	}

	if cfg.MinIOAutoCreateBucket {
		if err := storage.ensureBucket(ctx); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

func (s *MinIOStorage) Save(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("open upload stream: %w", err)
	}
	defer file.Close()

	objectKey := path.Join(utils.NewUUID(), sanitizeFilename(fileHeader.Filename))
	contentType := fileHeader.Header.Get("Content-Type")
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	thresholdBytes := s.multipartThresholdMB * 1024 * 1024
	if fileHeader.Size > thresholdBytes {
		if err := s.multipartUpload(ctx, objectKey, file, contentType); err != nil {
			return "", err
		}
		return objectKey, nil
	}

	body, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read upload payload: %w", err)
	}
	if err := s.putObject(ctx, objectKey, body, contentType, nil); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (s *MinIOStorage) Close() error { return nil }

func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	status, _, err := s.doSignedRequest(ctx, http.MethodHead, s.bucket, "", nil, nil, "")
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if status == http.StatusOK {
		return nil
	}
	if status != http.StatusNotFound {
		return fmt.Errorf("bucket check failed with status %d", status)
	}
	status, _, err = s.doSignedRequest(ctx, http.MethodPut, s.bucket, "", nil, nil, "")
	if err != nil {
		return fmt.Errorf("create minio bucket: %w", err)
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("create bucket failed with status %d", status)
	}
	return nil
}

func (s *MinIOStorage) putObject(ctx context.Context, objectKey string, body []byte, contentType string, query url.Values) error {
	status, responseBody, err := s.doSignedRequest(ctx, http.MethodPut, s.bucket, objectKey, body, query, contentType)
	if err != nil {
		return fmt.Errorf("upload object to minio: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("upload object to minio failed with status %d: %s", status, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func (s *MinIOStorage) multipartUpload(ctx context.Context, objectKey string, file multipart.File, contentType string) error {
	status, responseBody, err := s.doSignedRequest(ctx, http.MethodPost, s.bucket, objectKey, nil, url.Values{"uploads": []string{""}}, contentType)
	if err != nil {
		return fmt.Errorf("initiate multipart upload: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("initiate multipart upload failed with status %d: %s", status, strings.TrimSpace(string(responseBody)))
	}

	var initResponse initiateMultipartUploadResult
	if err := xml.Unmarshal(responseBody, &initResponse); err != nil {
		return fmt.Errorf("decode initiate multipart upload response: %w", err)
	}
	if strings.TrimSpace(initResponse.UploadID) == "" {
		return fmt.Errorf("multipart upload id missing from minio response")
	}

	partSize := int(s.multipartPartSizeMB * 1024 * 1024)
	if partSize <= 0 {
		partSize = 8 * 1024 * 1024
	}

	completed := make([]completeMultipartPart, 0)
	partNumber := 1
	for {
		buffer := make([]byte, partSize)
		n, readErr := io.ReadFull(file, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return fmt.Errorf("read multipart chunk: %w", readErr)
		}
		if n == 0 {
			break
		}
		partBody := buffer[:n]
		query := url.Values{
			"partNumber": []string{strconv.Itoa(partNumber)},
			"uploadId":   []string{initResponse.UploadID},
		}
		status, _, headers, err := s.doSignedRequestWithHeaders(ctx, http.MethodPut, s.bucket, objectKey, partBody, query, contentType)
		if err != nil {
			return fmt.Errorf("upload multipart chunk %d: %w", partNumber, err)
		}
		if status != http.StatusOK {
			return fmt.Errorf("upload multipart chunk %d failed with status %d", partNumber, status)
		}
		completed = append(completed, completeMultipartPart{PartNumber: partNumber, ETag: headers.Get("ETag")})
		partNumber++
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
	}

	completePayload, err := xml.Marshal(completeMultipartUpload{Parts: completed})
	if err != nil {
		return fmt.Errorf("encode multipart complete payload: %w", err)
	}
	status, responseBody, err = s.doSignedRequest(ctx, http.MethodPost, s.bucket, objectKey, completePayload, url.Values{"uploadId": []string{initResponse.UploadID}}, "application/xml")
	if err != nil {
		return fmt.Errorf("complete multipart upload: %w", err)
	}
	if status != http.StatusOK {
		return fmt.Errorf("complete multipart upload failed with status %d: %s", status, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func (s *MinIOStorage) doSignedRequest(ctx context.Context, method, bucket, objectKey string, body []byte, query url.Values, contentType string) (int, []byte, error) {
	status, responseBody, _, err := s.doSignedRequestWithHeaders(ctx, method, bucket, objectKey, body, query, contentType)
	return status, responseBody, err
}

func (s *MinIOStorage) doSignedRequestWithHeaders(ctx context.Context, method, bucket, objectKey string, body []byte, query url.Values, contentType string) (int, []byte, http.Header, error) {
	endpointURL := &url.URL{
		Scheme: s.scheme(),
		Host:   s.endpoint,
		Path:   "/" + strings.TrimPrefix(path.Join(bucket, objectKey), "/"),
	}
	if query != nil {
		endpointURL.RawQuery = canonicalQueryString(query)
	}

	payloadHash := sha256Hex(body)
	request, err := http.NewRequestWithContext(ctx, method, endpointURL.String(), bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}
	request.Header.Set("Host", s.endpoint)
	request.Header.Set("x-amz-content-sha256", payloadHash)
	request.Header.Set("x-amz-date", time.Now().UTC().Format("20060102T150405Z"))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	request.Header.Set("Authorization", s.authorizationHeader(request, payloadHash))

	response, err := s.httpClient.Do(request)
	if err != nil {
		return 0, nil, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return response.StatusCode, responseBody, response.Header, nil
}

func (s *MinIOStorage) authorizationHeader(request *http.Request, payloadHash string) string {
	amzDate := request.Header.Get("x-amz-date")
	date := amzDate[:8]
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{
		request.Method,
		request.URL.EscapedPath(),
		request.URL.RawQuery,
		"host:" + request.Host + "\n" +
			"x-amz-content-sha256:" + payloadHash + "\n" +
			"x-amz-date:" + amzDate + "\n",
		signedHeaders,
		payloadHash,
	}, "\n")
	credentialScope := date + "/" + defaultMinIORegion + "/s3/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := signingKey(s.secretKey, date, defaultMinIORegion, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s", s.accessKey, credentialScope, signedHeaders, signature)
}

func (s *MinIOStorage) scheme() string {
	if s.useSSL {
		return "https"
	}
	return "http"
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err == nil {
			return hex.EncodeToString(buf) + ".bin"
		}
		return "upload.bin"
	}
	return base
}

func canonicalQueryString(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		entries := append([]string(nil), values[key]...)
		sort.Strings(entries)
		for _, entry := range entries {
			parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(entry))
		}
	}
	return strings.ReplaceAll(strings.Join(parts, "&"), "+", "%20")
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key, payload []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(payload)
	return h.Sum(nil)
}

func signingKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}
