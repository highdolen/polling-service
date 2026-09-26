package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var voteScript = redis.NewScript(`
local dedup_key = KEYS[1]

if redis.call("EXISTS", dedup_key) == 1 then
	return 0
end

redis.call("SET", dedup_key, "1", "PX", ARGV[1])

for i = 2, #KEYS do
	redis.call("INCR", KEYS[i])
end

return 1
`)

type VoteStorage struct {
	client *redis.Client
}

func NewVoteStorage(client *redis.Client) *VoteStorage {
	return &VoteStorage{
		client: client,
	}
}

func (s *VoteStorage) Vote(
	ctx context.Context,
	pollID int64,
	optionIDs []int64,
	clientID string,
	ttl time.Duration,
) (bool, error) {
	keys := make([]string, 0, len(optionIDs)+1)

	keys = append(
		keys,
		voteKey(pollID, clientID),
	)

	for _, optionID := range optionIDs {
		keys = append(
			keys,
			resultKey(pollID, optionID),
		)
	}

	result, err := voteScript.Run(
		ctx,
		s.client,
		keys,
		ttl.Milliseconds(),
	).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (s *VoteStorage) GetResult(
	ctx context.Context,
	pollID int64,
	optionID int64,
) (int64, error) {
	value, err := s.client.Get(
		ctx,
		resultKey(pollID, optionID),
	).Int64()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}

		return 0, err
	}

	return value, nil
}

func voteKey(pollID int64, clientID string) string {
	return "vote:" + formatInt64(pollID) + ":" + clientID
}

func resultKey(pollID int64, optionID int64) string {
	return "result:" +
		formatInt64(pollID) +
		":" +
		formatInt64(optionID)
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}
