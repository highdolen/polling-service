package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"test_task/internal/handler"
	"test_task/internal/model"
	"test_task/internal/service"
)

func TestAdminHandler_CreatePoll_Success(t *testing.T) {
	pollService := &mockAdminPollService{
		poll: &model.Poll{
			ID:       1,
			Question: "Какой вариант выбираете?",
			Type:     "single",
			Status:   "draft",
			StartsAt: time.Date(
				2026,
				9,
				25,
				12,
				0,
				0,
				0,
				time.UTC,
			),
			EndsAt: time.Date(
				2026,
				9,
				25,
				12,
				1,
				0,
				0,
				time.UTC,
			),
		},
	}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	body := `{
		"question": "Какой вариант выбираете?",
		"type": "single",
		"starts_at": "2026-09-25T12:00:00Z",
		"ends_at": "2026-09-25T12:01:00Z",
		"options": [
			"Вариант A",
			"Вариант B"
		]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if !pollService.called {
		t.Fatal("expected poll service to be called")
	}

	if pollService.createdPoll == nil {
		t.Fatal("expected created poll to be set")
	}

	if pollService.createdPoll.Question != "Какой вариант выбираете?" {
		t.Fatalf(
			"unexpected question %q",
			pollService.createdPoll.Question,
		)
	}

	if pollService.createdPoll.Type != "single" {
		t.Fatalf(
			"unexpected poll type %q",
			pollService.createdPoll.Type,
		)
	}

	if len(pollService.createdOptions) != 2 {
		t.Fatalf(
			"expected 2 options, got %d",
			len(pollService.createdOptions),
		)
	}

	var response struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		Status   string         `json:"status"`
		Options  []model.Option `json:"options"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != 1 {
		t.Fatalf(
			"expected id 1, got %d",
			response.ID,
		)
	}

	if response.Question != "Какой вариант выбираете?" {
		t.Fatalf(
			"unexpected question %q",
			response.Question,
		)
	}

	if response.Status != "draft" {
		t.Fatalf(
			"expected status draft, got %q",
			response.Status,
		)
	}

	if len(response.Options) != 2 {
		t.Fatalf(
			"expected 2 options, got %d",
			len(response.Options),
		)
	}
}

func TestAdminHandler_CreatePoll_InvalidBody(t *testing.T) {
	pollService := &mockAdminPollService{}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(`invalid json`),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if pollService.called {
		t.Fatal("poll service should not be called")
	}
}

func TestAdminHandler_CreatePoll_InvalidStartsAt(t *testing.T) {
	pollService := &mockAdminPollService{}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	body := `{
		"question": "Question",
		"type": "single",
		"starts_at": "invalid",
		"ends_at": "2026-09-25T12:01:00Z",
		"options": ["A", "B"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if pollService.called {
		t.Fatal("poll service should not be called")
	}
}

func TestAdminHandler_CreatePoll_InvalidEndsAt(t *testing.T) {
	pollService := &mockAdminPollService{}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	body := `{
		"question": "Question",
		"type": "single",
		"starts_at": "2026-09-25T12:00:00Z",
		"ends_at": "invalid",
		"options": ["A", "B"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if pollService.called {
		t.Fatal("poll service should not be called")
	}
}

func TestAdminHandler_CreatePoll_InvalidQuestion(t *testing.T) {
	pollService := &mockAdminPollService{
		err: service.ErrInvalidQuestion,
	}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	body := `{
		"question": "",
		"type": "single",
		"starts_at": "2026-09-25T12:00:00Z",
		"ends_at": "2026-09-25T12:01:00Z",
		"options": ["A", "B"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestAdminHandler_CreatePoll_InvalidType(t *testing.T) {
	pollService := &mockAdminPollService{
		err: service.ErrInvalidPollType,
	}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	body := `{
		"question": "Question",
		"type": "unknown",
		"starts_at": "2026-09-25T12:00:00Z",
		"ends_at": "2026-09-25T12:01:00Z",
		"options": ["A", "B"]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/polls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreatePoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestAdminHandler_ListPolls_Success(t *testing.T) {
	polls := []model.Poll{
		{
			ID:       1,
			Question: "Question 1",
			Type:     "single",
			Status:   "draft",
		},
		{
			ID:       2,
			Question: "Question 2",
			Type:     "multiple",
			Status:   "active",
		},
	}

	pollService := &mockAdminPollService{
		polls: polls,
	}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls",
		nil,
	)

	rec := httptest.NewRecorder()

	h.ListPolls(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response []model.Poll

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(response) != 2 {
		t.Fatalf(
			"expected 2 polls, got %d",
			len(response),
		)
	}

	if response[0].ID != 1 {
		t.Fatalf(
			"expected first poll id 1, got %d",
			response[0].ID,
		)
	}
}

func TestAdminHandler_ListPolls_RepositoryError(t *testing.T) {
	pollService := &mockAdminPollService{
		err: errors.New("database unavailable"),
	}

	h := handler.NewAdminHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls",
		nil,
	)

	rec := httptest.NewRecorder()

	h.ListPolls(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestAdminHandler_GetResults_Success(t *testing.T) {
	resultService := &mockAdminResultService{
		results: []model.Result{
			{
				PollID:     1,
				OptionID:   1,
				VotesCount: 100,
			},
			{
				PollID:     1,
				OptionID:   2,
				VotesCount: 150,
			},
		},
	}

	h := handler.NewAdminHandler(
		nil,
		resultService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls/1/results",
		nil,
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.GetResults(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response []model.Result

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(response) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(response),
		)
	}

	if response[0].VotesCount != 100 {
		t.Fatalf(
			"expected 100 votes, got %d",
			response[0].VotesCount,
		)
	}

	if response[1].VotesCount != 150 {
		t.Fatalf(
			"expected 150 votes, got %d",
			response[1].VotesCount,
		)
	}
}

func TestAdminHandler_GetResults_NotFound(t *testing.T) {
	resultService := &mockAdminResultService{
		err: service.ErrPollNotFound,
	}

	h := handler.NewAdminHandler(
		nil,
		resultService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls/999/results",
		nil,
	)

	req.SetPathValue("id", "999")

	rec := httptest.NewRecorder()

	h.GetResults(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestAdminHandler_GetResults_InvalidID(t *testing.T) {
	resultService := &mockAdminResultService{}

	h := handler.NewAdminHandler(
		nil,
		resultService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls/abc/results",
		nil,
	)

	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()

	h.GetResults(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if resultService.called {
		t.Fatal("result service should not be called")
	}
}

func TestAdminHandler_GetResults_RepositoryError(t *testing.T) {
	resultService := &mockAdminResultService{
		err: errors.New("redis unavailable"),
	}

	h := handler.NewAdminHandler(
		nil,
		resultService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/polls/1/results",
		nil,
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.GetResults(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
