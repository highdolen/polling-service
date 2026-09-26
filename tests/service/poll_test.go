package service_test

import (
	"context"
	"testing"
	"time"

	"test_task/internal/model"
	"test_task/internal/service"
)

func TestPollService_Create(t *testing.T) {
	repo := newMockPollRepository()

	pollService := service.NewPollService(repo)

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

	pollService := service.NewPollService(repo)

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

	pollService := service.NewPollService(repo)

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

	pollService := service.NewPollService(repo)

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

	pollService := service.NewPollService(repo)

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

func TestPollService_Get(t *testing.T) {
	repo := newMockPollRepository()

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

	pollService := service.NewPollService(repo)

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
}

func TestPollService_Get_NotFound(t *testing.T) {
	repo := newMockPollRepository()

	pollService := service.NewPollService(repo)

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
