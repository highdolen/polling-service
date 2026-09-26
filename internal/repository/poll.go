package repository

import (
	"context"

	"test_task/internal/model"
)

type PollRepository interface {
	Create(ctx context.Context, poll *model.Poll) error
	GetByID(ctx context.Context, id int64) (*model.Poll, error)
	GetActive(ctx context.Context) (*model.Poll, error)
	List(ctx context.Context) ([]model.Poll, error)

	CreateOption(ctx context.Context, option *model.Option) error
	GetOptions(ctx context.Context, pollID int64) ([]model.Option, error)
	GetOptionByID(ctx context.Context, pollID, optionID int64) (*model.Option, error)
}
