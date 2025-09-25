package wrapper

import (
	"context"
	v1 "stream-server/api/v1"
	"stream-server/internal/interfaces"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type StreamUsecaseWrapper struct {
	uc interfaces.IUsecase
}

const UsecaseInstance = "StreamUsecase"

func NewStreamUsecaseWrapper(base interfaces.IUsecase) *StreamUsecaseWrapper {
	return &StreamUsecaseWrapper{uc: base}
}

func (s *StreamUsecaseWrapper) ListStreams(ctx context.Context, in *v1.ListStreamsRequest) (_ []*v1.Stream, err error) {
	ctx, span := otel.Tracer(UsecaseInstance).Start(ctx, "ListStreams")
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
	return s.uc.ListStreams(ctx, in)
}

func (s *StreamUsecaseWrapper) GetStream(ctx context.Context, in *v1.GetStreamRequest) (res *v1.Stream, err error) {
	ctx, span := otel.Tracer(UsecaseInstance).Start(ctx, "GetStream")
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
	return s.uc.GetStream(ctx, in)
}

func (s *StreamUsecaseWrapper) UpdateStream(ctx context.Context, in *v1.UpdateStreamRequest) (res *v1.Stream, err error) {
	ctx, span := otel.Tracer(UsecaseInstance).Start(ctx, "UpdateStream")
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
	return s.uc.UpdateStream(ctx, in)
}
