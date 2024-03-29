package repo

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func saveAliasToSeparateTable(ctx context.Context, tx pgx.Tx, alias string) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, "INSERT INTO aliases(name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", alias).Scan(&id); err == nil {
		return id, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	} else {
		return 0, err
	}
}

func saveTextToSeparateTable(ctx context.Context, tx pgx.Tx, text string, entities []byte) (int, error) {
	var (
		id  int
		err error
	)
	if entities == nil {
		entities = []byte("null")
	}
	if err = tx.QueryRow(ctx, "INSERT INTO texts(text, entities) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id", text, entities).Scan(&id); err == nil {
		return id, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	} else {
		return 0, err
	}
}

func saveLocationToSeparateTable(ctx context.Context, tx pgx.Tx, latitude, longitude float64) (int, error) {
	var id int
	if err := tx.QueryRow(ctx, "INSERT INTO locations(latitude, longitude) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id", latitude, longitude).Scan(&id); err == nil {
		return id, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
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
