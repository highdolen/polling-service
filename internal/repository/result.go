package repository

import (
	"context"

	"test_task/internal/model"
)

type ResultRepository interface {
	Create(ctx context.Context, result *model.Result) error
	GetByPollID(ctx context.Context, pollID int64) ([]model.Result, error)
	UpdateVotesCount(ctx context.Context, pollID, optionID, votesCount int64) error
}
