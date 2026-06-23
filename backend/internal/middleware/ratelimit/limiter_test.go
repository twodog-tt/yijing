package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wangxintong/yijing/backend/internal/middleware/ratelimit"
)

func TestLimiterAllowBySessionKey(t *testing.T) {
	limiter := ratelimit.NewLimiter(2)

	if !limiter.Allow("session:abc") {
		t.Fatal("first request should be allowed")
	}
	if !limiter.Allow("session:abc") {
		t.Fatal("second request should be allowed")
	}
	if limiter.Allow("session:abc") {
		t.Fatal("third request should be blocked")
	}
	if !limiter.Allow("session:xyz") {
		t.Fatal("different session should be allowed")
	}
}

func TestMiddlewareUsesSessionKeyFromBody(t *testing.T) {
	limiter := ratelimit.NewLimiter(1)
	mw := ratelimit.Middleware(limiter, true)

	called := 0
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		called++
	})

	body := `{"session_key":"user-1","question":"test"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/divinations", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	handler(rec1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/divinations", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler(rec2, req2)

	if called != 1 {
		t.Fatalf("expected handler called once, got %d", called)
	}
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}
