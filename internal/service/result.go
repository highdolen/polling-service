package service

import (
	"context"

	"polling-service/internal/model"
	"polling-service/internal/repository"
)

type ResultService struct {
	polls   repository.PollRepository
	results repository.ResultReader
}

func NewResultService(
	polls repository.PollRepository,
	results repository.ResultReader,
) *ResultService {
	return &ResultService{
		polls:   polls,
		results: results,
	}
}

func (s *ResultService) GetResults(
	ctx context.Context,
	pollID int64,
) ([]model.Result, error) {
	poll, err := s.polls.GetByID(ctx, pollID)
	if err != nil {
		return nil, err
	}

	if poll == nil {
		return nil, ErrPollNotFound
	}

	options, err := s.polls.GetOptions(ctx, pollID)
	if err != nil {
		return nil, err
	}

	results := make([]model.Result, 0, len(options))

	for _, option := range options {
		votesCount, err := s.results.GetResult(
			ctx,
			pollID,
			option.ID,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, model.Result{
			PollID:     pollID,
			OptionID:   option.ID,
			VotesCount: votesCount,
		})
	}

	return results, nil
}
