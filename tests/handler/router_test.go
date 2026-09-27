package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polling-service/internal/handler"
	"polling-service/internal/middleware"
	"polling-service/internal/model"
)

func TestRouter_PublicGetPoll(t *testing.T) {
	pollService := &mockPollService{
		poll: &model.Poll{
			ID:       1,
			Question: "Test question",
			Type:     "single",
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Yes",
			},
			{
				ID:     2,
				PollID: 1,
				Text:   "No",
			},
		},
	}

	voteService := &mockVoteService{}

	publicHandler := handler.NewPublicHandler(
		pollService,
		voteService,
	)

	adminPollService := &mockAdminPollService{}
	adminResultService := &mockAdminResultService{}

	adminHandler := handler.NewAdminHandler(
		adminPollService,
		adminResultService,
	)

	rateLimiter := middleware.NewRateLimiter(
		10,
		time.Minute,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		"secret-token",
		rateLimiter,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/1",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRouter_AdminRequiresToken(t *testing.T) {
	publicHandler := handler.NewPublicHandler(
		&mockPollService{},
		&mockVoteService{},
	)

	adminHandler := handler.NewAdminHandler(
		&mockAdminPollService{},
		&mockAdminResultService{},
	)

	rateLimiter := middleware.NewRateLimiter(
		10,
		time.Minute,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		"secret-token",
		rateLimiter,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(`{}`),
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestRouter_AdminWithToken(t *testing.T) {
	adminPollService := &mockAdminPollService{
		poll: &model.Poll{
			ID:       1,
			Question: "Test question",
			Type:     "single",
			Status:   "draft",
		},
	}

	adminHandler := handler.NewAdminHandler(
		&mockAdminPollService{
			poll: &model.Poll{
				ID:       1,
				Question: "Test question",
				Type:     "single",
				Status:   "draft",
			},
		},
		&mockAdminResultService{},
	)

	publicHandler := handler.NewPublicHandler(
		&mockPollService{},
		&mockVoteService{},
	)

	rateLimiter := middleware.NewRateLimiter(
		10,
		time.Minute,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		"secret-token",
		rateLimiter,
	)

	body := `{
		"question": "Test question",
		"type": "single",
		"starts_at": "2026-09-27T12:00:00Z",
		"ends_at": "2026-09-27T13:00:00Z",
		"options": ["Yes", "No"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	req.Header.Set("X-Admin-Token", "secret-token")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	_ = adminPollService
}

func TestRouter_CORSOptions(t *testing.T) {
	publicHandler := handler.NewPublicHandler(
		&mockPollService{},
		&mockVoteService{},
	)

	adminHandler := handler.NewAdminHandler(
		&mockAdminPollService{},
		&mockAdminResultService{},
	)

	rateLimiter := middleware.NewRateLimiter(
		10,
		time.Minute,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		"secret-token",
		rateLimiter,
	)

	req := httptest.NewRequest(
		http.MethodOptions,
		"/api/v1/polls/1",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf(
			"expected Access-Control-Allow-Origin *, got %q",
			got,
		)
	}
}

func TestRouter_VoteRateLimit(t *testing.T) {
	voteService := &mockVoteService{
		err: errors.New("test error"),
	}

	publicHandler := handler.NewPublicHandler(
		&mockPollService{},
		voteService,
	)

	adminHandler := handler.NewAdminHandler(
		&mockAdminPollService{},
		&mockAdminResultService{},
	)

	rateLimiter := middleware.NewRateLimiter(
		1,
		time.Minute,
	)

	router := handler.NewRouter(
		publicHandler,
		adminHandler,
		"secret-token",
		rateLimiter,
	)

	body := `{
		"option_ids": [1]
	}`

	req1 := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(body),
	)

	rec1 := httptest.NewRecorder()

	router.ServeHTTP(rec1, req1)

	// Первый запрос должен дойти до handler.
	if rec1.Code == http.StatusTooManyRequests {
		t.Fatal("first request was unexpectedly rate limited")
	}

	req2 := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(body),
	)

	rec2 := httptest.NewRecorder()

	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected second request status %d, got %d",
			http.StatusTooManyRequests,
			rec2.Code,
		)
	}
}
