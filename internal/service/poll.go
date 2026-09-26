package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"test_task/internal/model"
	"test_task/internal/repository"
)

var (
	ErrInvalidQuestion  = errors.New("question is required")
	ErrInvalidPollType  = errors.New("invalid poll type")
	ErrInvalidOptions   = errors.New("poll must have at least two options")
	ErrInvalidOption    = errors.New("option text is required")
	ErrInvalidPollTime  = errors.New("end time must be after start time")
	ErrPollTimeConflict = errors.New("poll time conflicts with another poll")
)

type PollService struct {
	polls repository.PollRepository
}

func NewPollService(
	polls repository.PollRepository,
) *PollService {
	return &PollService{
		polls: polls,
	}
}

func (s *PollService) Create(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
) error {
	if poll == nil {
		return ErrInvalidQuestion
	}

	if strings.TrimSpace(poll.Question) == "" {
		return ErrInvalidQuestion
	}

	if poll.Type != "single" && poll.Type != "multiple" {
		return ErrInvalidPollType
	}

	if len(options) < 2 {
		return ErrInvalidOptions
	}

	if !poll.EndsAt.After(poll.StartsAt) {
		return ErrInvalidPollTime
	}

	polls, err := s.polls.List(ctx)
	if err != nil {
		return err
	}

	for _, existingPoll := range polls {
		if poll.StartsAt.Before(existingPoll.EndsAt) &&
			poll.EndsAt.After(existingPoll.StartsAt) {
			return ErrPollTimeConflict
		}
	}

	for _, option := range options {
		if strings.TrimSpace(option.Text) == "" {
			return ErrInvalidOption
		}
	}

	if err := s.polls.Create(ctx, poll); err != nil {
		return err
	}

	for i := range options {
		options[i].PollID = poll.ID

		if err := s.polls.CreateOption(ctx, &options[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *PollService) Get(
	ctx context.Context,
	id int64,
) (*model.Poll, []model.Option, error) {
	poll, err := s.polls.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	if poll == nil {
		return nil, nil, ErrPollNotFound
	}

	updatePollStatus(poll)

	options, err := s.polls.GetOptions(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return poll, options, nil
}

func (s *PollService) GetActive(
	ctx context.Context,
) (*model.Poll, []model.Option, error) {
	poll, err := s.polls.GetActive(ctx)
	if err != nil {
		return nil, nil, err
	}

	if poll == nil {
		return nil, nil, ErrPollNotFound
	}

	poll.Status = "active"

	options, err := s.polls.GetOptions(ctx, poll.ID)
	if err != nil {
		return nil, nil, err
	}

	return poll, options, nil
}

func (s *PollService) List(
	ctx context.Context,
) ([]model.Poll, error) {
	polls, err := s.polls.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range polls {
		updatePollStatus(&polls[i])
	}

	return polls, nil
}

func updatePollStatus(poll *model.Poll) {
	now := time.Now()

	switch {
	case now.Before(poll.StartsAt):
		poll.Status = "draft"

	case now.Before(poll.EndsAt):
		poll.Status = "active"

	default:
		poll.Status = "finished"
	}
}
