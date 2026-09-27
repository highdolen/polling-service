package repository

import (
	"context"
	"errors"
	"time"

	"test_task/internal/model"
)

var (
	ErrCacheMiss        = errors.New("cache miss")
	ErrPollTimeConflict = errors.New("poll time conflict")
)

type PollRepository interface {
	Create(ctx context.Context, poll *model.Poll) error

	CreateWithOptions(
		ctx context.Context,
		poll *model.Poll,
		options []model.Option,
	) error

	GetByID(ctx context.Context, id int64) (*model.Poll, error)
	GetActive(ctx context.Context) (*model.Poll, error)
	List(ctx context.Context) ([]model.Poll, error)

	CreateOption(ctx context.Context, option *model.Option) error
	GetOptions(ctx context.Context, pollID int64) ([]model.Option, error)
	GetOptionByID(
		ctx context.Context,
		pollID,
		optionID int64,
	) (*model.Option, error)
}

type PollCache interface {
	Set(
		ctx context.Context,
		poll *model.Poll,
		options []model.Option,
		ttl time.Duration,
	) error

	Get(
		ctx context.Context,
		pollID int64,
	) (*model.Poll, []model.Option, error)

	Delete(
		ctx context.Context,
		pollID int64,
	) error

	GetActive(
		ctx context.Context,
	) (*model.Poll, []model.Option, error)

	SetActive(
		ctx context.Context,
		poll *model.Poll,
		options []model.Option,
		ttl time.Duration,
	) error
}
