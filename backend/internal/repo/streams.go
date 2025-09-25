package repo

import (
	"context"
	"fmt"

	"stream-server/internal/data/repo"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *StreamRepo) ListStreams(ctx context.Context) ([]repo.ListStreamsRow, error) {
	return r.queries.ListStreams(ctx)
}

func (r *StreamRepo) GetStream(ctx context.Context, ID pgtype.UUID) (repo.GetStreamRow, error) {
	return r.queries.GetStream(ctx, ID)
}

func (r *StreamRepo) UpdateStream(ctx context.Context, in repo.UpdateStreamParams) (res repo.UpdateStreamRow, err error) {
	tx, err := r.data.DBClientPool.Begin(ctx)
	if err != nil {
		return res, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.queries.WithTx(tx)

	res, err = qtx.UpdateStream(ctx, in)
	if err != nil {
		return res, fmt.Errorf("update stream: %w", err)
	}

	return res, tx.Commit(ctx)
}
