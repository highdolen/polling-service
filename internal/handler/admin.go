package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"test_task/internal/model"
	"test_task/internal/service"
)

type adminPollService interface {
	Create(
		ctx context.Context,
		poll *model.Poll,
		options []model.Option,
	) error

	List(
		ctx context.Context,
	) ([]model.Poll, error)
}

type adminResultService interface {
	GetResults(
		ctx context.Context,
		pollID int64,
	) ([]model.Result, error)
}

type AdminHandler struct {
	polls   adminPollService
	results adminResultService
}

func NewAdminHandler(
	polls adminPollService,
	results adminResultService,
) *AdminHandler {
	return &AdminHandler{
		polls:   polls,
		results: results,
	}
}

type createPollRequest struct {
	Question string   `json:"question"`
	Type     string   `json:"type"`
	StartsAt string   `json:"starts_at"`
	EndsAt   string   `json:"ends_at"`
	Options  []string `json:"options"`
}

func (h *AdminHandler) CreatePoll(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request createPollRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		http.Error(
			w,
			"invalid starts_at",
			http.StatusBadRequest,
		)
		return
	}

	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		http.Error(
			w,
			"invalid ends_at",
			http.StatusBadRequest,
		)
		return
	}

	options := make([]model.Option, 0, len(request.Options))

	for _, optionText := range request.Options {
		options = append(options, model.Option{
			Text: optionText,
		})
	}

	poll := &model.Poll{
		Question: request.Question,
		Type:     request.Type,
		Status:   "draft",
		StartsAt: startsAt,
		EndsAt:   endsAt,
	}

	if err := h.polls.Create(
		r.Context(),
		poll,
		options,
	); err != nil {
		h.handleCreatePollError(w, err)
		return
	}

	response := struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		Status   string         `json:"status"`
		StartsAt time.Time      `json:"starts_at"`
		EndsAt   time.Time      `json:"ends_at"`
		Options  []model.Option `json:"options"`
	}{
		ID:       poll.ID,
		Question: poll.Question,
		Type:     poll.Type,
		Status:   poll.Status,
		StartsAt: poll.StartsAt,
		EndsAt:   poll.EndsAt,
		Options:  options,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *AdminHandler) ListPolls(
	w http.ResponseWriter,
	r *http.Request,
) {
	polls, err := h.polls.List(r.Context())
	if err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(polls); err != nil {
		return
	}
}

func (h *AdminHandler) GetResults(
	w http.ResponseWriter,
	r *http.Request,
) {
	pollID, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil {
		http.Error(
			w,
			"invalid poll id",
			http.StatusBadRequest,
		)
		return
	}

	results, err := h.results.GetResults(
		r.Context(),
		pollID,
	)
	if err != nil {
		if errors.Is(err, service.ErrPollNotFound) {
			http.Error(
				w,
				"poll not found",
				http.StatusNotFound,
			)
			return
		}

		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		return
	}
}

func (h *AdminHandler) handleCreatePollError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidQuestion):
		http.Error(
			w,
			"question is required",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrInvalidPollType):
		http.Error(
			w,
			"invalid poll type",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrInvalidOptions):
		http.Error(
			w,
			"poll must have at least two options",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrInvalidOption):
		http.Error(
			w,
			"option text is required",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrInvalidPollTime):
		http.Error(
			w,
			"end time must be after start time",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrPollTimeConflict):
		http.Error(
			w,
			"poll time conflicts with another poll",
			http.StatusConflict,
		)

	default:
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}
