package repository

import "context"

type ResultReader interface {
	GetResult(
		ctx context.Context,
		pollID int64,
		optionID int64,
	) (int64, error)
}
