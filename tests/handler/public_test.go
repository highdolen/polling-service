package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"polling-service/internal/handler"
	"polling-service/internal/model"
	"polling-service/internal/service"
)

func TestPublicHandler_GetPoll_Success(t *testing.T) {
	poll := &model.Poll{
		ID:       1,
		Question: "Какой вариант выбираете?",
		Type:     "single",
		Status:   "active",
		StartsAt: time.Now().Add(-time.Minute),
		EndsAt:   time.Now().Add(time.Minute),
	}

	options := []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Вариант A",
		},
		{
			ID:     2,
			PollID: 1,
			Text:   "Вариант B",
		},
	}

	pollService := &mockPollService{
		poll:    poll,
		options: options,
	}

	h := handler.NewPublicHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/1",
		nil,
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.GetPoll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	var response struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		StartsAt time.Time      `json:"starts_at"`
		EndsAt   time.Time      `json:"ends_at"`
		Options  []model.Option `json:"options"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != poll.ID {
		t.Fatalf(
			"expected poll id %d, got %d",
			poll.ID,
			response.ID,
		)
	}

	if response.Question != poll.Question {
		t.Fatalf(
			"expected question %q, got %q",
			poll.Question,
			response.Question,
		)
	}

	if response.Type != poll.Type {
		t.Fatalf(
			"expected type %q, got %q",
			poll.Type,
			response.Type,
		)
	}

	if len(response.Options) != len(options) {
		t.Fatalf(
			"expected %d options, got %d",
			len(options),
			len(response.Options),
		)
	}

	if response.Options[0].ID != options[0].ID {
		t.Fatalf(
			"expected first option id %d, got %d",
			options[0].ID,
			response.Options[0].ID,
		)
	}

	if response.Options[0].Text != options[0].Text {
		t.Fatalf(
			"expected first option text %q, got %q",
			options[0].Text,
			response.Options[0].Text,
		)
	}

}

func TestPublicHandler_GetActivePoll_Success(t *testing.T) {
	now := time.Now()

	poll := &model.Poll{
		ID:       1,
		Question: "What is your favorite language?",
		Type:     "single",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	options := []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Go",
		},
		{
			ID:     2,
			PollID: 1,
			Text:   "Python",
		},
	}

	pollService := &mockPollService{
		poll:    poll,
		options: options,
	}

	voteService := &mockVoteService{}

	handler := handler.NewPublicHandler(
		pollService,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/active",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetActivePoll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		Options  []model.Option `json:"options"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Fatalf("expected poll ID 1, got %d", response.ID)
	}

	if response.Question != poll.Question {
		t.Fatalf(
			"expected question %q, got %q",
			poll.Question,
			response.Question,
		)
	}

	if response.Type != poll.Type {
		t.Fatalf(
			"expected type %q, got %q",
			poll.Type,
			response.Type,
		)
	}

	if len(response.Options) != 2 {
		t.Fatalf(
			"expected 2 options, got %d",
			len(response.Options),
		)
	}

	if response.Options[0].Text != "Go" {
		t.Fatalf(
			"expected first option %q, got %q",
			"Go",
			response.Options[0].Text,
		)
	}
}

func TestPublicHandler_GetActivePoll_NotFound(t *testing.T) {
	pollService := &mockPollService{
		poll:    nil,
		options: nil,
		err:     nil,
	}

	voteService := &mockVoteService{}

	handler := handler.NewPublicHandler(
		pollService,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/active",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetActivePoll(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestPublicHandler_GetActivePoll_RepositoryError(t *testing.T) {
	pollService := &mockPollService{
		err: errors.New("database error"),
	}

	voteService := &mockVoteService{}

	handler := handler.NewPublicHandler(
		pollService,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/active",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetActivePoll(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestPublicHandler_GetPoll_NotFound(t *testing.T) {
	pollService := &mockPollService{
		err: service.ErrPollNotFound,
	}

	h := handler.NewPublicHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/999",
		nil,
	)

	req.SetPathValue("id", "999")

	rec := httptest.NewRecorder()

	h.GetPoll(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}

}

func TestPublicHandler_GetPoll_InvalidID(t *testing.T) {
	pollService := &mockPollService{}

	h := handler.NewPublicHandler(
		pollService,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/polls/abc",
		nil,
	)

	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()

	h.GetPoll(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_Success(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[2]}`),
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if !voteService.called {
		t.Fatal("expected vote service to be called")
	}

	if voteService.pollID != 1 {
		t.Fatalf(
			"expected poll id 1, got %d",
			voteService.pollID,
		)
	}

	if len(voteService.optionIDs) != 1 {
		t.Fatalf(
			"expected 1 option id, got %d",
			len(voteService.optionIDs),
		)
	}

	if voteService.optionIDs[0] != 2 {
		t.Fatalf(
			"expected option id 2, got %d",
			voteService.optionIDs[0],
		)
	}

	cookies := rec.Result().Cookies()

	if len(cookies) != 1 {
		t.Fatalf(
			"expected 1 cookie, got %d",
			len(cookies),
		)
	}

	if cookies[0].Name != "client_id" {
		t.Fatalf(
			"expected cookie name client_id, got %q",
			cookies[0].Name,
		)
	}

	if cookies[0].Value == "" {
		t.Fatal("expected client_id cookie value to be set")
	}

	if voteService.clientID != cookies[0].Value {
		t.Fatalf(
			"expected client id %q, got %q",
			cookies[0].Value,
			voteService.clientID,
		)
	}

}

func TestPublicHandler_Vote_ExistingClient(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[2]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if voteService.clientID != "test-client-id" {
		t.Fatalf(
			"expected client id %q, got %q",
			"test-client-id",
			voteService.clientID,
		)
	}

	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("did not expect a new client_id cookie")
	}

}

func TestPublicHandler_Vote_MultipleOptions(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[1,2,3]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	expected := []int64{1, 2, 3}

	if len(voteService.optionIDs) != len(expected) {
		t.Fatalf(
			"expected %d option ids, got %d",
			len(expected),
			len(voteService.optionIDs),
		)
	}

	for i := range expected {
		if voteService.optionIDs[i] != expected[i] {
			t.Fatalf(
				"expected option id %d at index %d, got %d",
				expected[i],
				i,
				voteService.optionIDs[i],
			)
		}
	}

}

func TestPublicHandler_Vote_InvalidPollID(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/abc/vote",
		strings.NewReader(`{"option_ids":[1]}`),
	)

	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if voteService.called {
		t.Fatal("vote service should not be called")
	}

}

func TestPublicHandler_Vote_InvalidBody(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`invalid json`),
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if voteService.called {
		t.Fatal("vote service should not be called")
	}

}

func TestPublicHandler_Vote_NoOptions(t *testing.T) {
	voteService := &mockVoteService{}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[]}`),
	)

	req.SetPathValue("id", "1")

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if voteService.called {
		t.Fatal("vote service should not be called")
	}

}

func TestPublicHandler_Vote_AlreadyVoted(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrAlreadyVoted,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[1]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_PollNotFound(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrPollNotFound,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/999/vote",
		strings.NewReader(`{"option_ids":[1]}`),
	)

	req.SetPathValue("id", "999")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_OptionNotFound(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrOptionNotFound,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[999]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_PollFinished(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrPollFinished,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[1]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_PollNotStarted(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrPollNotStarted,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[1]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

}

func TestPublicHandler_Vote_InvalidVoteOptions(t *testing.T) {
	voteService := &mockVoteService{
		err: service.ErrInvalidVoteOptions,
	}

	h := handler.NewPublicHandler(
		nil,
		voteService,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/polls/1/vote",
		strings.NewReader(`{"option_ids":[1,2]}`),
	)

	req.SetPathValue("id", "1")

	req.AddCookie(&http.Cookie{
		Name:  "client_id",
		Value: "test-client-id",
	})

	rec := httptest.NewRecorder()

	h.Vote(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

}
