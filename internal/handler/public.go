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

	"github.com/google/uuid"
)

type pollService interface {
	Get(
		ctx context.Context,
		id int64,
	) (*model.Poll, []model.Option, error)

	GetActive(
		ctx context.Context,
	) (*model.Poll, []model.Option, error)
}

type voteService interface {
	Vote(
		ctx context.Context,
		pollID int64,
		optionIDs []int64,
		clientID string,
	) error
}

type PublicHandler struct {
	polls pollService
	votes voteService
}

func NewPublicHandler(
	polls pollService,
	votes voteService,
) *PublicHandler {
	return &PublicHandler{
		polls: polls,
		votes: votes,
	}
}

func (h *PublicHandler) GetPoll(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(
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

	poll, options, err := h.polls.Get(
		r.Context(),
		id,
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

	response := struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		StartsAt time.Time      `json:"starts_at"`
		EndsAt   time.Time      `json:"ends_at"`
		Options  []model.Option `json:"options"`
	}{
		ID:       poll.ID,
		Question: poll.Question,
		Type:     poll.Type,
		StartsAt: poll.StartsAt,
		EndsAt:   poll.EndsAt,
		Options:  options,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
	}
}

func (h *PublicHandler) Vote(
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

	var request struct {
		OptionIDs []int64 `json:"option_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if len(request.OptionIDs) == 0 {
		http.Error(
			w,
			"no options selected",
			http.StatusBadRequest)
		return
	}

	clientID, err := r.Cookie("client_id")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			http.Error(
				w,
				"failed to read client id",
				http.StatusInternalServerError,
			)
			return
		}

		clientIDValue := generateClientID()

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "client_id",
				Value:    clientIDValue,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
		)

		clientIDValueToUse := clientIDValue

		if err := h.votes.Vote(
			r.Context(),
			pollID,
			request.OptionIDs,
			clientIDValueToUse,
		); err != nil {
			h.handleVoteError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	if err := h.votes.Vote(
		r.Context(),
		pollID,
		request.OptionIDs,
		clientID.Value,
	); err != nil {
		h.handleVoteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PublicHandler) handleVoteError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrPollNotFound):
		http.Error(
			w,
			"poll not found",
			http.StatusNotFound,
		)

	case errors.Is(err, service.ErrOptionNotFound):
		http.Error(
			w,
			"option not found",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrNoOptions):
		http.Error(
			w,
			"no options selected",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrDuplicateOption):
		http.Error(
			w,
			"duplicate option",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrInvalidVoteOptions):
		http.Error(
			w,
			"invalid options for poll type",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrPollNotStarted):
		http.Error(
			w,
			"poll has not started",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrPollFinished):
		http.Error(
			w,
			"poll has finished",
			http.StatusBadRequest,
		)

	case errors.Is(err, service.ErrAlreadyVoted):
		http.Error(
			w,
			"already voted",
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

func (h *PublicHandler) GetActivePoll(
	w http.ResponseWriter,
	r *http.Request,
) {
	poll, options, err := h.polls.GetActive(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrPollNotFound) {
			http.Error(
				w,
				"no active poll",
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

	response := struct {
		ID       int64          `json:"id"`
		Question string         `json:"question"`
		Type     string         `json:"type"`
		StartsAt time.Time      `json:"starts_at"`
		EndsAt   time.Time      `json:"ends_at"`
		Options  []model.Option `json:"options"`
	}{
		ID:       poll.ID,
		Question: poll.Question,
		Type:     poll.Type,
		StartsAt: poll.StartsAt,
		EndsAt:   poll.EndsAt,
		Options:  options,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func generateClientID() string {
	return uuid.NewString()
}
