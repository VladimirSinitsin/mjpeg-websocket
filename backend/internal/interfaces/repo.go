package interfaces

import (
	"context"
	"stream-server/internal/data/repo"

	"github.com/jackc/pgx/v5/pgtype"
)

type IRepo interface {
	ListStreams(ctx context.Context) ([]repo.ListStreamsRow, error)
	GetStream(ctx context.Context, ID pgtype.UUID) (repo.GetStreamRow, error)
	UpdateStream(ctx context.Context, in repo.UpdateStreamParams) (res repo.UpdateStreamRow, err error)
}
