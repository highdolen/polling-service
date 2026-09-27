package integration

import (
	"context"
	"testing"
	"time"

	"test_task/internal/model"
	postgresstorage "test_task/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreatePollWithOptions(t *testing.T) {
	ctx := context.Background()

	db := postgresClient(t)

	repository := postgresstorage.NewPollRepository(db)

	now := time.Now().Add(24 * time.Hour)

	poll := &model.Poll{
		Question: "Какой язык программирования вам нравится?",
		Type:     "single",
		Status:   "draft",
		StartsAt: now,
		EndsAt:   now.Add(time.Hour),
	}

	options := []model.Option{
		{
			Text: "Go",
		},
		{
			Text: "Python",
		},
		{
			Text: "Java",
		},
	}

	err := repository.CreateWithOptions(
		ctx,
		poll,
		options,
	)
	if err != nil {
		db.Close()

		t.Fatalf(
			"failed to create poll with options: %v",
			err,
		)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			ctx,
			"DELETE FROM polls WHERE id = $1",
			poll.ID,
		)
		if err != nil {
			t.Errorf(
				"failed to clean test data: %v",
				err,
			)
		}

		db.Close()
	})

	if poll.ID == 0 {
		t.Fatal("expected poll ID to be set")
	}

	savedPoll, err := repository.GetByID(
		ctx,
		poll.ID,
	)
	if err != nil {
		t.Fatalf("failed to get poll: %v", err)
	}

	if savedPoll == nil {
		t.Fatal("expected poll to exist")
	}

	if savedPoll.Question != poll.Question {
		t.Fatalf(
			"expected question %q, got %q",
			poll.Question,
			savedPoll.Question,
		)
	}

	savedOptions, err := repository.GetOptions(
		ctx,
		poll.ID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get poll options: %v",
			err,
		)
	}

	if len(savedOptions) != len(options) {
		t.Fatalf(
			"expected %d options, got %d",
			len(options),
			len(savedOptions),
		)
	}

	for i := range savedOptions {
		if savedOptions[i].Text != options[i].Text {
			t.Fatalf(
				"expected option %q, got %q",
				options[i].Text,
				savedOptions[i].Text,
			)
		}

		if savedOptions[i].PollID != poll.ID {
			t.Fatalf(
				"expected option poll ID %d, got %d",
				poll.ID,
				savedOptions[i].PollID,
			)
		}
	}
}

func postgresClient(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	dsn := "postgres://postgres:postgres@localhost:5432/polling?sslmode=disable"

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf(
			"failed to create postgres pool: %v",
			err,
		)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()

		t.Fatalf(
			"postgres is unavailable: %v",
			err,
		)
	}

	return db
}
