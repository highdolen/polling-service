package repository

import (
	"context"
	"time"
)

type VoteRepository interface {
	Vote(
		ctx context.Context,
		pollID int64,
		optionIDs []int64,
		clientID string,
		ttl time.Duration,
	) (bool, error)
}
