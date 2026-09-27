package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"polling-service/internal/model"
	resultsync "polling-service/internal/sync"
)

type mockPollRepository struct {
	polls   []model.Poll
	options map[int64][]model.Option

	listErr       error
	getOptionsErr error
}

func (m *mockPollRepository) List(
	ctx context.Context,
) ([]model.Poll, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	return m.polls, nil
}

func (m *mockPollRepository) GetOptions(
	ctx context.Context,
	pollID int64,
) ([]model.Option, error) {
	if m.getOptionsErr != nil {
		return nil, m.getOptionsErr
	}

	return m.options[pollID], nil
}

type mockResultReader struct {
	results map[int64]int64
	err     error
}

func (m *mockResultReader) GetResult(
	ctx context.Context,
	pollID int64,
	optionID int64,
) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	return m.results[optionID], nil
}

type updatedResult struct {
	pollID     int64
	optionID   int64
	votesCount int64
}

type mockResultRepository struct {
	updates []updatedResult
	err     error
}

func (m *mockResultRepository) UpdateVotesCount(
	ctx context.Context,
	pollID int64,
	optionID int64,
	votesCount int64,
) error {
	if m.err != nil {
		return m.err
	}

	m.updates = append(m.updates, updatedResult{
		pollID:     pollID,
		optionID:   optionID,
		votesCount: votesCount,
	})

	return nil
}

func TestResultSync_Sync(t *testing.T) {
	pollRepository := &mockPollRepository{
		polls: []model.Poll{
			{
				ID: 1,
			},
		},
		options: map[int64][]model.Option{
			1: {
				{
					ID:     1,
					PollID: 1,
					Text:   "Go",
				},
				{
					ID:     2,
					PollID: 1,
					Text:   "Python",
				},
			},
		},
	}

	resultReader := &mockResultReader{
		results: map[int64]int64{
			1: 10,
			2: 5,
		},
	}

	resultRepository := &mockResultRepository{}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	err := sync.Sync(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resultRepository.updates) != 2 {
		t.Fatalf(
			"expected 2 updates, got %d",
			len(resultRepository.updates),
		)
	}

	if resultRepository.updates[0].pollID != 1 {
		t.Fatalf(
			"expected poll ID 1, got %d",
			resultRepository.updates[0].pollID,
		)
	}

	if resultRepository.updates[0].optionID != 1 {
		t.Fatalf(
			"expected option ID 1, got %d",
			resultRepository.updates[0].optionID,
		)
	}

	if resultRepository.updates[0].votesCount != 10 {
		t.Fatalf(
			"expected 10 votes, got %d",
			resultRepository.updates[0].votesCount,
		)
	}

	if resultRepository.updates[1].optionID != 2 {
		t.Fatalf(
			"expected option ID 2, got %d",
			resultRepository.updates[1].optionID,
		)
	}

	if resultRepository.updates[1].votesCount != 5 {
		t.Fatalf(
			"expected 5 votes, got %d",
			resultRepository.updates[1].votesCount,
		)
	}
}

func TestResultSync_Sync_RepeatDoesNotIncreaseVotes(t *testing.T) {
	pollRepository := &mockPollRepository{
		polls: []model.Poll{
			{
				ID: 1,
			},
		},
		options: map[int64][]model.Option{
			1: {
				{
					ID:     1,
					PollID: 1,
					Text:   "Go",
				},
			},
		},
	}

	resultReader := &mockResultReader{
		results: map[int64]int64{
			1: 10,
		},
	}

	resultRepository := &mockResultRepository{}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	ctx := context.Background()

	if err := sync.Sync(ctx); err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	if err := sync.Sync(ctx); err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	if len(resultRepository.updates) != 2 {
		t.Fatalf(
			"expected 2 updates, got %d",
			len(resultRepository.updates),
		)
	}

	if resultRepository.updates[0].votesCount != 10 {
		t.Fatalf(
			"expected first sync to write 10 votes, got %d",
			resultRepository.updates[0].votesCount,
		)
	}

	if resultRepository.updates[1].votesCount != 10 {
		t.Fatalf(
			"expected second sync to write 10 votes, got %d",
			resultRepository.updates[1].votesCount,
		)
	}
}

func TestResultSync_Sync_ListError(t *testing.T) {
	expectedErr := errors.New("list polls failed")

	pollRepository := &mockPollRepository{
		listErr: expectedErr,
	}

	resultReader := &mockResultReader{}
	resultRepository := &mockResultRepository{}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	err := sync.Sync(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}

	if len(resultRepository.updates) != 0 {
		t.Fatalf(
			"expected no updates, got %d",
			len(resultRepository.updates),
		)
	}
}

func TestResultSync_Sync_GetOptionsError(t *testing.T) {
	expectedErr := errors.New("get options failed")

	pollRepository := &mockPollRepository{
		polls: []model.Poll{
			{
				ID: 1,
			},
		},
		options:       map[int64][]model.Option{},
		getOptionsErr: expectedErr,
	}

	resultReader := &mockResultReader{}
	resultRepository := &mockResultRepository{}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	err := sync.Sync(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestResultSync_Sync_GetResultError(t *testing.T) {
	expectedErr := errors.New("get result failed")

	pollRepository := &mockPollRepository{
		polls: []model.Poll{
			{
				ID: 1,
			},
		},
		options: map[int64][]model.Option{
			1: {
				{
					ID:     1,
					PollID: 1,
				},
			},
		},
	}

	resultReader := &mockResultReader{
		err: expectedErr,
	}

	resultRepository := &mockResultRepository{}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	err := sync.Sync(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}

	if len(resultRepository.updates) != 0 {
		t.Fatalf(
			"expected no updates, got %d",
			len(resultRepository.updates),
		)
	}
}

func TestResultSync_Sync_UpdateError(t *testing.T) {
	expectedErr := errors.New("update result failed")

	pollRepository := &mockPollRepository{
		polls: []model.Poll{
			{
				ID: 1,
			},
		},
		options: map[int64][]model.Option{
			1: {
				{
					ID:     1,
					PollID: 1,
				},
			},
		},
	}

	resultReader := &mockResultReader{
		results: map[int64]int64{
			1: 10,
		},
	}

	resultRepository := &mockResultRepository{
		err: expectedErr,
	}

	sync := resultsync.NewResultSync(
		pollRepository,
		resultReader,
		resultRepository,
		time.Second,
	)

	err := sync.Sync(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}
}
