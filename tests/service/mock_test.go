package service_test

import (
	"context"

	"polling-service/internal/model"
)

type mockPollRepository struct {
	polls   map[int64]*model.Poll
	options map[int64][]model.Option
}

func newMockPollRepository() *mockPollRepository {
	return &mockPollRepository{
		polls:   make(map[int64]*model.Poll),
		options: make(map[int64][]model.Option),
	}
}

func (m *mockPollRepository) Create(
	ctx context.Context,
	poll *model.Poll,
) error {
	poll.ID = int64(len(m.polls) + 1)
	m.polls[poll.ID] = poll

	return nil
}

func (m *mockPollRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Poll, error) {
	poll, ok := m.polls[id]
	if !ok {
		return nil, nil
	}

	return poll, nil
}

func (m *mockPollRepository) GetActive(
	ctx context.Context,
) (*model.Poll, error) {
	for _, poll := range m.polls {
		if poll.Status == "active" {
			return poll, nil
		}
	}

	return nil, nil
}

func (m *mockPollRepository) List(
	ctx context.Context,
) ([]model.Poll, error) {
	result := make([]model.Poll, 0, len(m.polls))

	for _, poll := range m.polls {
		result = append(result, *poll)
	}

	return result, nil
}

func (m *mockPollRepository) CreateOption(
	ctx context.Context,
	option *model.Option,
) error {
	m.options[option.PollID] = append(
		m.options[option.PollID],
		*option,
	)

	return nil
}

func (m *mockPollRepository) GetOptions(
	ctx context.Context,
	pollID int64,
) ([]model.Option, error) {
	return m.options[pollID], nil
}

func (m *mockPollRepository) GetOptionByID(
	ctx context.Context,
	pollID int64,
	optionID int64,
) (*model.Option, error) {
	for _, option := range m.options[pollID] {
		if option.ID == optionID {
			return &option, nil
		}
	}

	return nil, nil
}
