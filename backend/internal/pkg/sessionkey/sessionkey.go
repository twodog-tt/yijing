package sessionkey

import (
	"errors"
	"net/http"
	"strings"
)

const (
	HeaderName = "X-Session-Key"
	MaxLength  = 64
)

var (
	ErrConflict = errors.New("session_key conflict between header and body")
	ErrTooLong  = errors.New("session_key exceeds max length")
)

// FromHeader reads the anonymous session key from the standard request header.
func FromHeader(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(HeaderName))
}

// FromQuery reads session_key from URL query. Analysis APIs must not use this.
func FromQuery(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("session_key"))
}

// Resolve prefers header over body value when they are compatible.
func Resolve(headerValue, bodyValue string) string {
	key, err := ResolveForCreate(headerValue, bodyValue)
	if err != nil {
		return ""
	}
	return key
}

// ResolveForCreate returns the session key for create requests.
// If both header and body provide different non-empty keys, ErrConflict is returned.
func ResolveForCreate(headerValue, bodyValue string) (string, error) {
	header := strings.TrimSpace(headerValue)
	body := strings.TrimSpace(bodyValue)
	if header != "" && body != "" && header != body {
		return "", ErrConflict
	}
	key := header
	if key == "" {
		key = body
	}
	if err := ValidateLength(key); err != nil {
		return "", err
	}
	return key, nil
}

// ValidateLength checks session_key length against storage limits.
// Empty key is allowed here; callers decide whether empty is valid.
func ValidateLength(key string) error {
	if len(key) > MaxLength {
		return ErrTooLong
	}
	return nil
}
