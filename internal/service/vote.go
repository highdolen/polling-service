package service

import (
	"context"
	"errors"
	"time"

	"test_task/internal/repository"
)

var (
	ErrPollNotFound       = errors.New("poll not found")
	ErrOptionNotFound     = errors.New("option not found")
	ErrPollNotStarted     = errors.New("poll has not started")
	ErrPollFinished       = errors.New("poll has finished")
	ErrAlreadyVoted       = errors.New("already voted")
	ErrNoOptions          = errors.New("no options selected")
	ErrDuplicateOption    = errors.New("duplicate option")
	ErrInvalidVoteOptions = errors.New("invalid options for poll type")
)

type VoteService struct {
	polls repository.PollRepository
	votes repository.VoteRepository
}

func NewVoteService(
	polls repository.PollRepository,
	votes repository.VoteRepository,
) *VoteService {
	return &VoteService{
		polls: polls,
		votes: votes,
	}
}

func (s *VoteService) Vote(
	ctx context.Context,
	pollID int64,
	optionIDs []int64,
	clientID string,
) error {
	if len(optionIDs) == 0 {
		return ErrNoOptions
	}

	poll, err := s.polls.GetByID(ctx, pollID)
	if err != nil {
		return err
	}

	if poll == nil {
		return ErrPollNotFound
	}

	now := time.Now()

	if now.Before(poll.StartsAt) {
		return ErrPollNotStarted
	}

	if now.After(poll.EndsAt) {
		return ErrPollFinished
	}

	if poll.Type == "single" && len(optionIDs) != 1 {
		return ErrInvalidVoteOptions
	}

	seen := make(map[int64]struct{}, len(optionIDs))

	for _, optionID := range optionIDs {
		if optionID <= 0 {
			return ErrOptionNotFound
		}

		if _, exists := seen[optionID]; exists {
			return ErrDuplicateOption
		}

		seen[optionID] = struct{}{}

		option, err := s.polls.GetOptionByID(
			ctx,
			pollID,
			optionID,
		)
		if err != nil {
			return err
		}

		if option == nil {
			return ErrOptionNotFound
		}
	}

	ttl := time.Until(poll.EndsAt)

	if ttl <= 0 {
		return ErrPollFinished
	}

	ok, err := s.votes.Vote(
		ctx,
		pollID,
		optionIDs,
		clientID,
		ttl,
	)
	if err != nil {
		return err
	}

	if !ok {
		return ErrAlreadyVoted
	}

	return nil
}
