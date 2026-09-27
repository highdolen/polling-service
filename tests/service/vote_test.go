package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"test_task/internal/model"
	"test_task/internal/repository"
	"test_task/internal/service"
)

type mockVoteRepository struct {
	success   bool
	called    bool
	err       error
	optionIDs []int64
	clientID  string
	ttl       time.Duration
}

func (m *mockVoteRepository) Vote(
	ctx context.Context,
	pollID int64,
	optionIDs []int64,
	clientID string,
	ttl time.Duration,
) (bool, error) {
	m.called = true
	m.optionIDs = optionIDs
	m.clientID = clientID
	m.ttl = ttl

	if m.err != nil {
		return false, m.err
	}

	return m.success, nil
}

type mockVotePollCache struct {
	polls map[int64]cachedVotePoll
}

type cachedVotePoll struct {
	poll    model.Poll
	options []model.Option
}

func newMockVotePollCache() *mockVotePollCache {
	return &mockVotePollCache{
		polls: make(map[int64]cachedVotePoll),
	}
}

func (m *mockVotePollCache) Set(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
	ttl time.Duration,
) error {
	m.polls[poll.ID] = cachedVotePoll{
		poll:    *poll,
		options: options,
	}

	return nil
}

func (m *mockVotePollCache) Get(
	ctx context.Context,
	pollID int64,
) (*model.Poll, []model.Option, error) {
	cached, ok := m.polls[pollID]
	if !ok {
		return nil, nil, repository.ErrCacheMiss
	}

	return &cached.poll, cached.options, nil
}

func (m *mockVotePollCache) Delete(
	ctx context.Context,
	pollID int64,
) error {
	delete(m.polls, pollID)

	return nil
}

func (m *mockVotePollCache) GetActive(
	ctx context.Context,
) (*model.Poll, []model.Option, error) {
	return nil, nil, repository.ErrCacheMiss
}

func (m *mockVotePollCache) SetActive(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
	ttl time.Duration,
) error {
	return nil
}

func TestVoteService_Vote_Success(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "multiple",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
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
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
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
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: false,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "finished",
			StartsAt: now.Add(-2 * time.Minute),
			EndsAt:   now.Add(-time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	expectedErr := errors.New("redis is unavailable")

	voteRepo := &mockVoteRepository{
		err: expectedErr,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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

func TestVoteService_Vote_CacheMiss(t *testing.T) {
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

	cache := newMockVotePollCache()

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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

	if _, ok := cache.polls[1]; !ok {
		t.Fatal("expected poll to be saved in cache")
	}
}

func TestVoteService_Vote_NoOptions(t *testing.T) {
	pollRepo := newMockPollRepository()
	cache := newMockVotePollCache()
	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		nil,
		"test-client",
	)

	if !errors.Is(err, service.ErrNoOptions) {
		t.Fatalf(
			"expected ErrNoOptions, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}
}

func TestVoteService_Vote_DuplicateOption(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "multiple",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   now.Add(time.Minute),
		},
		options: []model.Option{
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
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1, 1},
		"test-client",
	)

	if !errors.Is(err, service.ErrDuplicateOption) {
		t.Fatalf(
			"expected ErrDuplicateOption, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}
}

func TestVoteService_Vote_PollNotStarted(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()
	cache := newMockVotePollCache()

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "draft",
			StartsAt: now.Add(time.Minute),
			EndsAt:   now.Add(2 * time.Minute),
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
		voteRepo,
	)

	err := voteService.Vote(
		context.Background(),
		1,
		[]int64{1},
		"test-client",
	)

	if !errors.Is(err, service.ErrPollNotStarted) {
		t.Fatalf(
			"expected ErrPollNotStarted, got %v",
			err,
		)
	}

	if voteRepo.called {
		t.Fatal("expected VoteRepository.Vote not to be called")
	}
}

func TestVoteService_Vote_PassesTTL(t *testing.T) {
	now := time.Now()

	pollRepo := newMockPollRepository()
	cache := newMockVotePollCache()

	endsAt := now.Add(2 * time.Minute)

	cache.polls[1] = cachedVotePoll{
		poll: model.Poll{
			ID:       1,
			Question: "Test question?",
			Type:     "single",
			Status:   "active",
			StartsAt: now.Add(-time.Minute),
			EndsAt:   endsAt,
		},
		options: []model.Option{
			{
				ID:     1,
				PollID: 1,
				Text:   "Option 1",
			},
		},
	}

	voteRepo := &mockVoteRepository{
		success: true,
	}

	voteService := service.NewVoteService(
		pollRepo,
		cache,
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

	if voteRepo.clientID != "test-client" {
		t.Fatalf(
			"expected client ID %q, got %q",
			"test-client",
			voteRepo.clientID,
		)
	}

	if len(voteRepo.optionIDs) != 1 ||
		voteRepo.optionIDs[0] != 1 {
		t.Fatalf(
			"expected option IDs [1], got %v",
			voteRepo.optionIDs,
		)
	}

	if voteRepo.ttl <= 0 {
		t.Fatal("expected positive TTL")
	}

	if voteRepo.ttl > 2*time.Minute {
		t.Fatalf(
			"expected TTL to be no greater than 2 minutes, got %v",
			voteRepo.ttl,
		)
	}
}
