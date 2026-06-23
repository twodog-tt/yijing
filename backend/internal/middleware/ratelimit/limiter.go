package ratelimit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wangxintong/yijing/backend/internal/pkg/response"
)

type Limiter struct {
	mu           sync.Mutex
	maxPerMinute int
	hits         map[string][]time.Time
}

func NewLimiter(maxPerMinute int) *Limiter {
	if maxPerMinute < 1 {
		maxPerMinute = 1
	}
	return &Limiter{
		maxPerMinute: maxPerMinute,
		hits:         make(map[string][]time.Time),
	}
}

func (l *Limiter) Allow(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	l.mu.Lock()
	defer l.mu.Unlock()

	times := l.hits[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= l.maxPerMinute {
		l.hits[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	l.hits[key] = filtered
	return true
}

func Middleware(limiter *Limiter, enabled bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !enabled {
				next(w, r)
				return
			}

			key, restored, err := extractKey(r)
			if err != nil {
				response.Error(w, http.StatusBadRequest, response.CodeBadRequest, "invalid json body")
				return
			}
			r.Body = restored

			if !limiter.Allow(key) {
				response.Error(w, http.StatusTooManyRequests, response.CodeTooManyRequests, "操作太频繁，请稍后再试")
				return
			}

			next(w, r)
		}
	}
}

func extractKey(r *http.Request) (string, io.ReadCloser, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", r.Body, err
	}
	_ = r.Body.Close()

	var payload struct {
		SessionKey string `json:"session_key"`
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", io.NopCloser(bytes.NewReader(body)), err
		}
	}

	key := strings.TrimSpace(payload.SessionKey)
	if key == "" {
		key = "ip:" + clientIP(r)
	} else {
		key = "session:" + key
	}

	return key, io.NopCloser(bytes.NewReader(body)), nil
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		return host[:idx]
	}
	return host
}
