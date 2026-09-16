package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txCtxKey struct{}
type actorCtxKey struct{}

// WithTx runs fn inside a database transaction. If ctx already carries a
// transaction (nested call), that transaction is reused and no new
// BEGIN/COMMIT is issued, so services can freely compose within one another
// while still guaranteeing a single atomic unit of work at the outermost call.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx) error) error {
	if tx, ok := TxFromContext(ctx); ok {
		return fn(ctx, tx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, txCtxKey{}, tx)

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			return fmt.Errorf("%w (rollback also failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
	return tx, ok
}

// Q returns the active Querier for this context: the in-flight transaction
// if present, otherwise the bare pool. Repositories should call this instead
// of holding their own pool reference so they transparently participate in
// whatever transaction the calling service started.
func Q(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return pool
}

func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, actor)
}

func ActorFromContext(ctx context.Context) string {
	if a, ok := ctx.Value(actorCtxKey{}).(string); ok {
		return a
	}
	return "system"
}
