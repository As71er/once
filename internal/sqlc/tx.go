package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

func WithTx[T any](ctx context.Context, db *sql.DB, fn func(q *Queries) (T, error)) (T, error) {
	var entity T

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return entity, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := fn(New(tx))
	if err != nil {
		return entity, fmt.Errorf("exec transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return entity, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}
