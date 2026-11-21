// Package postgres implements the gateway repositories against PostgreSQL,
// through pgx. Every repository must pass the contract in
// infrastructure/repository/repositorytest.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// querier is what a repository runs SQL through: a *pgxpool.Pool outside a
// transaction, or the pgx.Tx TxManager.WithinTx put in the context.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

// querierFrom returns the pgx.Tx a WithinTx call put in ctx, or pool when
// there is none: a call outside WithinTx still runs, just not atomically
// with any other.
func querierFrom(ctx context.Context, pool querier) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}

// isUniqueViolation reports whether err is a violation of the named unique
// index or constraint (PostgreSQL error code 23505).
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}
