package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"test_task/internal/model"
)

type PollStorage struct {
	client *redis.Client
}

func NewPollStorage(client *redis.Client) *PollStorage {
	return &PollStorage{
		client: client,
	}
}

type cachedPoll struct {
	Poll    model.Poll
	Options []model.Option
}

func (s *PollStorage) Set(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
	ttl time.Duration,
) error {
	data, err := json.Marshal(cachedPoll{
		Poll:    *poll,
		Options: options,
	})
	if err != nil {
		return err
	}

	return s.client.Set(
		ctx,
		pollKey(poll.ID),
		data,
		ttl,
	).Err()
}

func (s *PollStorage) Get(
	ctx context.Context,
	pollID int64,
) (*model.Poll, []model.Option, error) {
	data, err := s.client.Get(
		ctx,
		pollKey(pollID),
	).Bytes()
	if err != nil {
		return nil, nil, err
	}

	var cached cachedPoll

	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, nil, err
	}

	return &cached.Poll, cached.Options, nil
}

func (s *PollStorage) Delete(
	ctx context.Context,
	pollID int64,
) error {
	return s.client.Del(
		ctx,
		pollKey(pollID),
	).Err()
}

func pollKey(pollID int64) string {
	return "poll:" + formatInt64(pollID)
}
