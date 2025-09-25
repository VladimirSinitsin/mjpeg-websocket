package wrapper

import (
	"context"
	"stream-server/internal/data/repo"
	"stream-server/internal/interfaces"

	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type StreamRepoWrapper struct {
	repo interfaces.IRepo
}

const RepoInstance = "StreamRepo"

func NewStreamRepoWrapper(repo interfaces.IRepo) *StreamRepoWrapper {
	return &StreamRepoWrapper{repo: repo}
}

func (s *StreamRepoWrapper) ListStreams(ctx context.Context) (_ []repo.ListStreamsRow, err error) {
	ctx, span := otel.Tracer(RepoInstance).Start(ctx, "ListStreams")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.repo.ListStreams(ctx)
}

func (s *StreamRepoWrapper) GetStream(ctx context.Context, ID pgtype.UUID) (res repo.GetStreamRow, err error) {
	ctx, span := otel.Tracer(RepoInstance).Start(ctx, "GetStream")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.repo.GetStream(ctx, ID)
}

func (s *StreamRepoWrapper) UpdateStream(ctx context.Context, in repo.UpdateStreamParams) (res repo.UpdateStreamRow, err error) {
	ctx, span := otel.Tracer(RepoInstance).Start(ctx, "UpdateStream")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.repo.UpdateStream(ctx, in)
}
