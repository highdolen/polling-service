package sync

import (
	"context"
	"log"
	"time"

	"test_task/internal/model"
)

type pollRepository interface {
	List(ctx context.Context) ([]model.Poll, error)
	GetOptions(ctx context.Context, pollID int64) ([]model.Option, error)
}

type resultReader interface {
	GetResult(ctx context.Context, pollID int64, optionID int64) (int64, error)
}

type resultRepository interface {
	UpdateVotesCount(
		ctx context.Context,
		pollID int64,
		optionID int64,
		votesCount int64,
	) error
}

type ResultSync struct {
	polls    pollRepository
	results  resultReader
	storage  resultRepository
	interval time.Duration
}

func NewResultSync(
	polls pollRepository,
	results resultReader,
	storage resultRepository,
	interval time.Duration,
) *ResultSync {
	return &ResultSync{
		polls:    polls,
		results:  results,
		storage:  storage,
		interval: interval,
	}
}

func (s *ResultSync) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Sync(ctx); err != nil {
				log.Printf("result sync failed: %v", err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (s *ResultSync) Sync(ctx context.Context) error {
	polls, err := s.polls.List(ctx)
	if err != nil {
		return err
	}

	for _, poll := range polls {
		if err := s.syncPoll(ctx, poll.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *ResultSync) syncPoll(
	ctx context.Context,
	pollID int64,
) error {
	options, err := s.polls.GetOptions(ctx, pollID)
	if err != nil {
		return err
	}

	for _, option := range options {
		votesCount, err := s.results.GetResult(
			ctx,
			pollID,
			option.ID,
		)
		if err != nil {
			return err
		}

		if err := s.storage.UpdateVotesCount(
			ctx,
			pollID,
			option.ID,
			votesCount,
		); err != nil {
			return err
		}
	}

	return nil
}
