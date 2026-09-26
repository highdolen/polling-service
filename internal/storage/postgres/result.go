package postgres

import (
	"context"

	"test_task/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ResultRepository struct {
	db *pgxpool.Pool
}

func NewResultRepository(db *pgxpool.Pool) *ResultRepository {
	return &ResultRepository{
		db: db,
	}
}

func (r *ResultRepository) Create(
	ctx context.Context,
	result *model.Result,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO poll_results (
			poll_id,
			option_id,
			votes_count
		)
		VALUES ($1, $2, $3)
		`,
		result.PollID,
		result.OptionID,
		result.VotesCount,
	)

	return err
}

func (r *ResultRepository) GetByPollID(
	ctx context.Context,
	pollID int64,
) ([]model.Result, error) {
	rows, err := r.db.Query(
		ctx,
		`
		SELECT poll_id, option_id, votes_count
		FROM poll_results
		WHERE poll_id = $1
		ORDER BY option_id
		`,
		pollID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.Result, 0)

	for rows.Next() {
		var result model.Result

		if err := rows.Scan(
			&result.PollID,
			&result.OptionID,
			&result.VotesCount,
		); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ResultRepository) UpdateVotesCount(
	ctx context.Context,
	pollID int64,
	optionID int64,
	votesCount int64,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO poll_results (
			poll_id,
			option_id,
			votes_count,
			updated_at
		)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (poll_id, option_id)
		DO UPDATE SET
			votes_count = EXCLUDED.votes_count,
			updated_at = NOW()
		`,
		pollID,
		optionID,
		votesCount,
	)

	return err
}
