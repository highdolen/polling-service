package service

import (
	"context"
	"errors"
	"time"

	"polling-service/internal/model"
	"polling-service/internal/repository"
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
	cache repository.PollCache
	votes repository.VoteRepository
}

func NewVoteService(
	polls repository.PollRepository,
	cache repository.PollCache,
	votes repository.VoteRepository,
) *VoteService {
	return &VoteService{
		polls: polls,
		cache: cache,
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

	poll, options, err := s.getPoll(ctx, pollID)
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

	if !now.Before(poll.EndsAt) {
		return ErrPollFinished
	}

	if poll.Type == "single" && len(optionIDs) != 1 {
		return ErrInvalidVoteOptions
	}

	optionSet := make(map[int64]struct{}, len(options))

	for _, option := range options {
		optionSet[option.ID] = struct{}{}
	}

	seen := make(map[int64]struct{}, len(optionIDs))

	for _, optionID := range optionIDs {
		if optionID <= 0 {
			return ErrOptionNotFound
		}

		if _, exists := seen[optionID]; exists {
			return ErrDuplicateOption
		}

		if _, exists := optionSet[optionID]; !exists {
			return ErrOptionNotFound
		}

		seen[optionID] = struct{}{}
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

func (s *VoteService) getPoll(
	ctx context.Context,
	pollID int64,
) (*model.Poll, []model.Option, error) {
	poll, options, err := s.cache.Get(ctx, pollID)
	if err == nil {
		return poll, options, nil
	}

	if !errors.Is(err, repository.ErrCacheMiss) {
		return nil, nil, err
	}

	poll, err = s.polls.GetByID(ctx, pollID)
	if err != nil {
		return nil, nil, err
	}

	if poll == nil {
		return nil, nil, nil
	}

	options, err = s.polls.GetOptions(ctx, pollID)
	if err != nil {
		return nil, nil, err
	}

	ttl := time.Until(poll.EndsAt)

	if ttl > 0 {
		if err := s.cache.Set(
			ctx,
			poll,
			options,
			ttl,
		); err != nil {
			return nil, nil, err
		}
	}

	return poll, options, nil
}
