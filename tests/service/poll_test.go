package service_test

import (
	"context"
	"testing"
	"time"

	"test_task/internal/model"
	"test_task/internal/repository"
	"test_task/internal/service"
)

type mockPollCache struct {
	polls map[int64]cachedPoll
}

type cachedPoll struct {
	poll    model.Poll
	options []model.Option
}

func newMockPollCache() *mockPollCache {
	return &mockPollCache{
		polls: make(map[int64]cachedPoll),
	}
}

func (m *mockPollCache) Set(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
	ttl time.Duration,
) error {
	m.polls[poll.ID] = cachedPoll{
		poll:    *poll,
		options: options,
	}

	return nil
}

func (m *mockPollCache) Get(
	ctx context.Context,
	pollID int64,
) (*model.Poll, []model.Option, error) {
	cached, ok := m.polls[pollID]
	if !ok {
		return nil, nil, repository.ErrCacheMiss
	}

	return &cached.poll, cached.options, nil
}

func (m *mockPollCache) Delete(
	ctx context.Context,
	pollID int64,
) error {
	delete(m.polls, pollID)

	return nil
}

func (m *mockPollCache) GetActive(
	ctx context.Context,
) (*model.Poll, []model.Option, error) {
	for _, cached := range m.polls {
		if cached.poll.Status == "active" {
			return &cached.poll, cached.options, nil
		}
	}

	return nil, nil, repository.ErrCacheMiss
}

func (m *mockPollCache) SetActive(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
	ttl time.Duration,
) error {
	m.polls[poll.ID] = cachedPoll{
		poll:    *poll,
		options: options,
	}

	return nil
}

// CreateWithOptions добавляет poll и все его options в mock repository.
func (m *mockPollRepository) CreateWithOptions(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
) error {
	poll.ID = int64(len(m.polls) + 1)

	m.polls[poll.ID] = poll

	for i := range options {
		options[i].ID = int64(len(m.options[poll.ID]) + 1)
		options[i].PollID = poll.ID

		m.options[poll.ID] = append(
			m.options[poll.ID],
			options[i],
		)
	}

	return nil
}

func TestPollService_Create(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	poll := &model.Poll{
		Question: "Какой язык программирования вам нравится?",
		Type:     "single",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}

	options := []model.Option{
		{
			Text: "Go",
		},
		{
			Text: "Java",
		},
		{
			Text: "Python",
		},
	}

	err := pollService.Create(
		context.Background(),
		poll,
		options,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if poll.ID == 0 {
		t.Fatal("expected poll ID to be set")
	}

	savedOptions := repo.options[poll.ID]

	if len(savedOptions) != 3 {
		t.Fatalf(
			"expected 3 options, got %d",
			len(savedOptions),
		)
	}
}

func TestPollService_Create_EmptyQuestion(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	poll := &model.Poll{
		Question: "",
		Type:     "single",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}

	options := []model.Option{
		{
			Text: "Да",
		},
		{
			Text: "Нет",
		},
	}

	err := pollService.Create(
		context.Background(),
		poll,
		options,
	)

	if err != service.ErrInvalidQuestion {
		t.Fatalf(
			"expected ErrInvalidQuestion, got %v",
			err,
		)
	}
}

func TestPollService_Create_NotEnoughOptions(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	poll := &model.Poll{
		Question: "Выберите вариант",
		Type:     "single",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}

	options := []model.Option{
		{
			Text: "Да",
		},
	}

	err := pollService.Create(
		context.Background(),
		poll,
		options,
	)

	if err != service.ErrInvalidOptions {
		t.Fatalf(
			"expected ErrInvalidOptions, got %v",
			err,
		)
	}
}

func TestPollService_Create_EmptyOption(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	poll := &model.Poll{
		Question: "Выберите вариант",
		Type:     "single",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}

	options := []model.Option{
		{
			Text: "Да",
		},
		{
			Text: " ",
		},
	}

	err := pollService.Create(
		context.Background(),
		poll,
		options,
	)

	if err != service.ErrInvalidOption {
		t.Fatalf(
			"expected ErrInvalidOption, got %v",
			err,
		)
	}
}

func TestPollService_Create_InvalidTime(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	now := time.Now()

	poll := &model.Poll{
		Question: "Выберите вариант",
		Type:     "single",
		StartsAt: now,
		EndsAt:   now,
	}

	options := []model.Option{
		{
			Text: "Да",
		},
		{
			Text: "Нет",
		},
	}

	err := pollService.Create(
		context.Background(),
		poll,
		options,
	)

	if err != service.ErrInvalidPollTime {
		t.Fatalf(
			"expected ErrInvalidPollTime, got %v",
			err,
		)
	}
}

func TestPollService_Create_PollTimeConflict(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	now := time.Now()

	firstPoll := &model.Poll{
		Question: "Первый опрос",
		Type:     "single",
		StartsAt: now,
		EndsAt:   now.Add(10 * time.Minute),
	}

	options := []model.Option{
		{
			Text: "Да",
		},
		{
			Text: "Нет",
		},
	}

	err := pollService.Create(
		context.Background(),
		firstPoll,
		options,
	)
	if err != nil {
		t.Fatalf("expected no error for first poll, got %v", err)
	}

	secondPoll := &model.Poll{
		Question: "Второй опрос",
		Type:     "single",
		StartsAt: now.Add(5 * time.Minute),
		EndsAt:   now.Add(15 * time.Minute),
	}

	err = pollService.Create(
		context.Background(),
		secondPoll,
		options,
	)

	if err != service.ErrPollTimeConflict {
		t.Fatalf(
			"expected ErrPollTimeConflict, got %v",
			err,
		)
	}
}

func TestPollService_Create_AdjacentPolls(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	now := time.Now()

	firstPoll := &model.Poll{
		Question: "Первый опрос",
		Type:     "single",
		StartsAt: now,
		EndsAt:   now.Add(10 * time.Minute),
	}

	options := []model.Option{
		{
			Text: "Да",
		},
		{
			Text: "Нет",
		},
	}

	err := pollService.Create(
		context.Background(),
		firstPoll,
		options,
	)
	if err != nil {
		t.Fatalf("expected no error for first poll, got %v", err)
	}

	secondPoll := &model.Poll{
		Question: "Второй опрос",
		Type:     "single",
		StartsAt: now.Add(10 * time.Minute),
		EndsAt:   now.Add(20 * time.Minute),
	}

	err = pollService.Create(
		context.Background(),
		secondPoll,
		options,
	)

	if err != nil {
		t.Fatalf(
			"expected adjacent polls to be allowed, got %v",
			err,
		)
	}
}

func TestPollService_Get(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	poll := &model.Poll{
		ID:       1,
		Question: "Какой язык программирования вам нравится?",
		Type:     "single",
		Status:   "active",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(time.Hour),
	}

	repo.polls[1] = poll

	repo.options[1] = []model.Option{
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
	}

	pollService := service.NewPollService(
		repo,
		cache,
	)

	gotPoll, gotOptions, err := pollService.Get(
		context.Background(),
		1,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotPoll == nil {
		t.Fatal("expected poll, got nil")
	}

	if gotPoll.ID != 1 {
		t.Fatalf(
			"expected poll ID 1, got %d",
			gotPoll.ID,
		)
	}

	if gotPoll.Question != poll.Question {
		t.Fatalf(
			"expected question %q, got %q",
			poll.Question,
			gotPoll.Question,
		)
	}

	if len(gotOptions) != 2 {
		t.Fatalf(
			"expected 2 options, got %d",
			len(gotOptions),
		)
	}

	if _, ok := cache.polls[1]; !ok {
		t.Fatal("expected poll to be saved in cache")
	}
}

func TestPollService_Get_NotFound(t *testing.T) {
	repo := newMockPollRepository()
	cache := newMockPollCache()

	pollService := service.NewPollService(
		repo,
		cache,
	)

	_, _, err := pollService.Get(
		context.Background(),
		999,
	)

	if err != service.ErrPollNotFound {
		t.Fatalf(
			"expected ErrPollNotFound, got %v",
			err,
		)
	}
}
