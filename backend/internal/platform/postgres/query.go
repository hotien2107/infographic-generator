package postgres

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func FormatQuery(query string, args ...any) (string, error) {
	formatted := query
	for index, arg := range args {
		placeholder := "$" + strconv.Itoa(index+1)
		literal, err := literal(arg)
		if err != nil {
			return "", fmt.Errorf("format arg %d: %w", index+1, err)
		}
		formatted = strings.ReplaceAll(formatted, placeholder, literal)
	}
	return formatted, nil
}

func literal(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "NULL", nil
	case string:
		return quoteString(v), nil
	case []byte:
		return quoteString(string(v)), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case bool:
		if v {
			return "TRUE", nil
		}
		return "FALSE", nil
	case time.Time:
		return quoteString(v.UTC().Format(time.RFC3339Nano)) + "::timestamptz", nil
	case *time.Time:
		if v == nil {
			return "NULL", nil
		}
		return literal(*v)
	case *string:
		if v == nil {
			return "NULL", nil
		}
		return quoteString(*v), nil
	default:
		if stringer, ok := value.(fmt.Stringer); ok {
			return quoteString(stringer.String()), nil
		}
		return quoteString(fmt.Sprintf("%v", value)), nil
	}
}

func quoteString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
