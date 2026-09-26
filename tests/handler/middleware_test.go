package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"test_task/internal/middleware"
)

func TestAdminAuth_Success(t *testing.T) {
	handler := middleware.AdminAuth("secret")(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	req.Header.Set(
		"X-Admin-Token",
		"secret",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestAdminAuth_InvalidToken(t *testing.T) {
	handler := middleware.AdminAuth("secret")(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	req.Header.Set(
		"X-Admin-Token",
		"wrong",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestAdminAuth_MissingToken(t *testing.T) {
	handler := middleware.AdminAuth("secret")(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestCORS(t *testing.T) {
	handler := middleware.CORS(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/1",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if got := rec.Header().Get(
		"Access-Control-Allow-Origin",
	); got != "*" {
		t.Fatalf(
			"expected CORS origin *, got %q",
			got,
		)
	}
}

func TestCORS_Options(t *testing.T) {
	handler := middleware.CORS(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			t.Fatal("next handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodOptions,
		"/api/v1/polls/1",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := middleware.NewRateLimiter(
		2,
		time.Second,
	)

	handler := limiter.Middleware(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/polls/1/vote",
			nil,
		)

		req.RemoteAddr = "127.0.0.1:12345"

		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: expected status %d, got %d",
				i+1,
				http.StatusOK,
				rec.Code,
			)
		}
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		nil,
	)

	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			rec.Code,
		)
	}
}

func TestRateLimiter_AllowsAfterWindow(t *testing.T) {
	limiter := middleware.NewRateLimiter(
		1,
		10*time.Millisecond,
	)

	handler := limiter.Middleware(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		nil,
	)

	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected first request status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	time.Sleep(20 * time.Millisecond)

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		nil,
	)

	req.RemoteAddr = "127.0.0.1:12345"

	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected request after window status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
