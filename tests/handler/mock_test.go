package handler_test

import (
	"context"

	"polling-service/internal/model"
)

type mockPollService struct {
	poll    *model.Poll
	options []model.Option
	err     error
}

func (m *mockPollService) Get(
	ctx context.Context,
	id int64,
) (*model.Poll, []model.Option, error) {
	return m.poll, m.options, m.err
}

func (m *mockPollService) GetActive(
	ctx context.Context,
) (*model.Poll, []model.Option, error) {
	return m.poll, m.options, m.err
}

type mockVoteService struct {
	err       error
	called    bool
	pollID    int64
	optionIDs []int64
	clientID  string
}

func (m *mockVoteService) Vote(
	ctx context.Context,
	pollID int64,
	optionIDs []int64,
	clientID string,
) error {
	m.called = true
	m.pollID = pollID
	m.optionIDs = optionIDs
	m.clientID = clientID

	return m.err

}

type mockAdminPollService struct {
	polls          []model.Poll
	poll           *model.Poll
	err            error
	called         bool
	createdPoll    *model.Poll
	createdOptions []model.Option
}

func (m *mockAdminPollService) Create(
	ctx context.Context,
	poll *model.Poll,
	options []model.Option,
) error {
	m.called = true
	m.createdPoll = poll
	m.createdOptions = options

	if m.poll != nil {
		*poll = *m.poll
	}

	return m.err

}

func (m *mockAdminPollService) List(
	ctx context.Context,
) ([]model.Poll, error) {
	return m.polls, m.err
}

type mockAdminResultService struct {
	results []model.Result
	err     error
	called  bool
}

func (m *mockAdminResultService) GetResults(
	ctx context.Context,
	pollID int64,
) ([]model.Result, error) {
	m.called = true

	return m.results, m.err

}
