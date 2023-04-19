package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func saveAliasToSeparateTable(ctx context.Context, tx pgx.Tx, alias string) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, "INSERT INTO aliases(name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", alias).Scan(&id); err == nil {
		return id, nil
	} else if err == pgx.ErrNoRows {
		return 0, nil
	} else {
		return 0, err
	}
}

func saveTextToSeparateTable(ctx context.Context, tx pgx.Tx, text string) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, "INSERT INTO texts(text) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", text).Scan(&id); err == nil {
		return id, nil
	} else if err == pgx.ErrNoRows {
		return 0, nil
	} else {
		return 0, err
	}
}

// rowsAffectedAdder is used to sum all values of [pgconn.CommandTag.RowsAffected] and return them as a single value.
type rowsAffectedAdder struct {
	rowsAffected int64
}

func (s *rowsAffectedAdder) RowsAffected() int64 {
	return s.rowsAffected
}

func (s *rowsAffectedAdder) Add(res pgconn.CommandTag) {
	s.rowsAffected += res.RowsAffected()
}
