package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"test_task/internal/model"
)

type PollRepository struct {
	db *pgxpool.Pool
}

func NewPollRepository(db *pgxpool.Pool) *PollRepository {
	return &PollRepository{
		db: db,
	}
}

func (r *PollRepository) Create(ctx context.Context, poll *model.Poll) error {
	query := `
		INSERT INTO polls (
			question,
			type,
			status,
			starts_at,
			ends_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		poll.Question,
		poll.Type,
		poll.Status,
		poll.StartsAt,
		poll.EndsAt,
	).Scan(
		&poll.ID,
		&poll.CreatedAt,
		&poll.UpdatedAt,
	)
}

func (r *PollRepository) GetByID(ctx context.Context, id int64) (*model.Poll, error) {
	query := `
		SELECT
			id,
			question,
			type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM polls
		WHERE id = $1
	`

	var poll model.Poll

	err := r.db.QueryRow(ctx, query, id).Scan(
		&poll.ID,
		&poll.Question,
		&poll.Type,
		&poll.Status,
		&poll.StartsAt,
		&poll.EndsAt,
		&poll.CreatedAt,
		&poll.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &poll, nil
}

func (r *PollRepository) List(ctx context.Context) ([]model.Poll, error) {
	query := `
		SELECT
			id,
			question,
			type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM polls
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	polls := make([]model.Poll, 0)

	for rows.Next() {
		var poll model.Poll

		if err := rows.Scan(
			&poll.ID,
			&poll.Question,
			&poll.Type,
			&poll.Status,
			&poll.StartsAt,
			&poll.EndsAt,
			&poll.CreatedAt,
			&poll.UpdatedAt,
		); err != nil {
			return nil, err
		}

		polls = append(polls, poll)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return polls, nil
}

func (r *PollRepository) CreateOption(ctx context.Context, option *model.Option) error {
	query := `
		INSERT INTO poll_options (
			poll_id,
			text
		)
		VALUES ($1, $2)
		RETURNING id
	`

	return r.db.QueryRow(
		ctx,
		query,
		option.PollID,
		option.Text,
	).Scan(&option.ID)
}

func (r *PollRepository) GetOptions(ctx context.Context, pollID int64) ([]model.Option, error) {
	query := `
		SELECT
			id,
			poll_id,
			text
		FROM poll_options
		WHERE poll_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]model.Option, 0)

	for rows.Next() {
		var option model.Option

		if err := rows.Scan(
			&option.ID,
			&option.PollID,
			&option.Text,
		); err != nil {
			return nil, err
		}

		options = append(options, option)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

func (r *PollRepository) GetOptionByID(
	ctx context.Context,
	pollID int64,
	optionID int64,
) (*model.Option, error) {
	query := `
		SELECT
			id,
			poll_id,
			text
		FROM poll_options
		WHERE poll_id = $1
		  AND id = $2
	`

	var option model.Option

	err := r.db.QueryRow(
		ctx,
		query,
		pollID,
		optionID,
	).Scan(
		&option.ID,
		&option.PollID,
		&option.Text,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &option, nil
}

func (r *PollRepository) GetActive(ctx context.Context) (*model.Poll, error) {
	const query = `
		SELECT
			id,
			question,
			type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM polls
		WHERE starts_at <= NOW()
		  AND ends_at > NOW()
		ORDER BY starts_at DESC
		LIMIT 1
	`

	var poll model.Poll

	err := r.db.QueryRow(
		ctx,
		query,
	).Scan(
		&poll.ID,
		&poll.Question,
		&poll.Type,
		&poll.Status,
		&poll.StartsAt,
		&poll.EndsAt,
		&poll.CreatedAt,
		&poll.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &poll, nil
}
