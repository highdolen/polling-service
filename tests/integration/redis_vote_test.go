package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	pollredis "polling-service/internal/storage/redis"

	goredis "github.com/redis/go-redis/v9"
)

func TestConcurrentVote(t *testing.T) {
	ctx := context.Background()

	client := redisClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("failed to close redis client: %v", err)
		}
	}()

	const (
		pollID   = 1
		optionID = 1
		clientID = "test-client"
	)

	err := client.Del(
		ctx,
		"vote:1:test-client",
		"result:1:1",
	).Err()
	if err != nil {
		t.Fatalf("failed to clean test data: %v", err)
	}

	storage := pollredis.NewVoteStorage(client)

	var (
		wg        sync.WaitGroup
		successes int
		mu        sync.Mutex
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			success, err := storage.Vote(
				ctx,
				pollID,
				[]int64{optionID},
				clientID,
				time.Minute,
			)
			if err != nil {
				t.Errorf("vote failed: %v", err)
				return
			}

			if success {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successes != 1 {
		t.Fatalf(
			"expected exactly 1 successful vote, got %d",
			successes,
		)
	}

	count, err := storage.GetResult(
		ctx,
		pollID,
		optionID,
	)
	if err != nil {
		t.Fatalf("get result failed: %v", err)
	}

	if count != 1 {
		t.Fatalf(
			"expected result counter to be 1, got %d",
			count,
		)
	}

	err = client.Del(
		ctx,
		"vote:1:test-client",
		"result:1:1",
	).Err()
	if err != nil {
		t.Fatalf("failed to clean test data after test: %v", err)
	}

}

func redisClient(t *testing.T) *goredis.Client {
	t.Helper()

	client := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			t.Fatalf(
				"redis is unavailable: %v; failed to close client: %v",
				err,
				closeErr,
			)
		}

		t.Fatalf("redis is unavailable: %v", err)
	}

	return client

}
