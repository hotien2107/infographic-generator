package postgres

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func FormatQuery(query string, args ...any) (string, error) {
	literals := make([]string, len(args))
	for index, arg := range args {
		literal, err := literal(arg)
		if err != nil {
			return "", fmt.Errorf("format arg %d: %w", index+1, err)
		}
		literals[index] = literal
	}

	var builder strings.Builder
	builder.Grow(len(query))

	for index := 0; index < len(query); index++ {
		if query[index] != '$' {
			builder.WriteByte(query[index])
			continue
		}

		end := index + 1
		for end < len(query) && unicode.IsDigit(rune(query[end])) {
			end++
		}
		if end == index+1 {
			builder.WriteByte(query[index])
			continue
		}

		placeholderIndex, err := strconv.Atoi(query[index+1 : end])
		if err != nil {
			return "", fmt.Errorf("parse placeholder %q: %w", query[index:end], err)
		}
		if placeholderIndex < 1 || placeholderIndex > len(literals) {
			builder.WriteString(query[index:end])
			index = end - 1
			continue
		}

		builder.WriteString(literals[placeholderIndex-1])
		index = end - 1
	}

	return builder.String(), nil
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
