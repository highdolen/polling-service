package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"test_task/internal/model"
	"test_task/internal/service"
)

type mockVoteRepository struct {
	success bool
	called  bool
	err     error
}

func (m *mockVoteRepository) Vote(
	ctx context.Context,
	pollID int64,
	optionIDs []int64,
	clientID string,
	ttl time.Duration,
) (bool, error) {
	m.called = true

	if m.err != nil {
		return false, m.err
	}

	return m.success, nil

}

func TestVoteService_Vote_Success(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1},
		"test-client",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !voteRepo.called {
		t.Fatal("expected VoteRepository.Vote to be called")
	}

}

func TestVoteService_Vote_MultipleOptions(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "multiple",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
		{
			ID:     2,
			PollID: 1,
			Text:   "Option 2",
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1, 2},
		"test-client",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !voteRepo.called {
		t.Fatal("expected VoteRepository.Vote to be called")
	}

}

func TestVoteService_Vote_SinglePollMultipleOptions(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
		{
			ID:     2,
			PollID: 1,
			Text:   "Option 2",
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1, 2},
		"test-client",
	)

	if !errors.Is(err, service.ErrInvalidVoteOptions) {
		t.Fatalf(
			"expected ErrInvalidVoteOptions, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}

}

func TestVoteService_Vote_AlreadyVoted(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
	}

	voteRepo := &mockVoteRepository{
		success: false,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1},
		"test-client",
	)

	if !errors.Is(err, service.ErrAlreadyVoted) {
		t.Fatalf(
			"expected ErrAlreadyVoted, got %v",
			err,
		)
	}

}

func TestVoteService_Vote_PollNotFound(t *testing.T) {
	pollRepo := newMockPollRepository()

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		999,
		[]int64{1},
		"test-client",
	)

	if !errors.Is(err, service.ErrPollNotFound) {
		t.Fatalf(
			"expected ErrPollNotFound, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}

}

func TestVoteService_Vote_OptionNotFound(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{999},
		"test-client",
	)

	if !errors.Is(err, service.ErrOptionNotFound) {
		t.Fatalf(
			"expected ErrOptionNotFound, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}

}

func TestVoteService_Vote_PollFinished(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "finished",
		StartsAt: now.Add(-2 * time.Minute),
		EndsAt:   now.Add(-time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1},
		"test-client",
	)

	if !errors.Is(err, service.ErrPollFinished) {
		t.Fatalf(
			"expected ErrPollFinished, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}

}

func TestVoteService_Vote_RepositoryError(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()

	pollRepo.polls[1] = &model.Poll{
		ID:       1,
		Question: "Test question?",
		Type:     "single",
		Status:   "active",
		StartsAt: now.Add(-time.Minute),
		EndsAt:   now.Add(time.Minute),
	}

	pollRepo.options[1] = []model.Option{
		{
			ID:     1,
			PollID: 1,
			Text:   "Option 1",
		},
	}

	expectedErr := errors.New("redis is unavailable")

	voteRepo := &mockVoteRepository{
		err: expectedErr,
	}

	voteService := service.NewVoteService(
		pollRepo,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1},
		"test-client",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}

	if !voteRepo.called {
		t.Fatal("expected VoteRepository.Vote to be called")
	}

}
