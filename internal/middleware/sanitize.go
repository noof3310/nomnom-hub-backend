package middleware

import "strings"

func mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// header keys to mask (lowercase)
var sensitiveHeaders = map[string]struct{}{
	"x-line-signature": {},
	"authorization":    {},
	"x-api-key":        {},
}

func isSensitiveHeader(key string) bool {
	_, ok := sensitiveHeaders[strings.ToLower(key)]
	return ok
}
