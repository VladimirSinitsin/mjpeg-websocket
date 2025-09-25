package wrapper

import (
	"context"
	"net/http"
	v1 "stream-server/api/v1"
	"stream-server/internal/interfaces"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type StreamServiceWrapper struct {
	v1.UnimplementedStreamServiceServer
	service interfaces.IStreamService
}

const StreamServiceInstance = "StreamService"

func NewStreamServiceWrapper(base interfaces.IStreamService) *StreamServiceWrapper {
	return &StreamServiceWrapper{service: base}
}

func (s *StreamServiceWrapper) StreamWSHandler() http.HandlerFunc {
	// TODO: подумоть
	return s.service.StreamWSHandler()
}

func (s *StreamServiceWrapper) ListStreams(ctx context.Context, in *v1.ListStreamsRequest) (res *v1.ListStreamsResponse, err error) {
	ctx, span := otel.Tracer(StreamServiceInstance).Start(ctx, "StreamService.ListStreams")
	defer func() {
		span.SetAttributes(
			attribute.Stringer("in", in),
			attribute.Stringer("res", res),
		)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.service.ListStreams(ctx, in)
}

func (s *StreamServiceWrapper) GetStream(ctx context.Context, in *v1.GetStreamRequest) (res *v1.GetStreamResponse, err error) {
	ctx, span := otel.Tracer(StreamServiceInstance).Start(ctx, "StreamService.GetStream")
	defer func() {
		span.SetAttributes(
			attribute.Stringer("in", in),
			attribute.Stringer("res", res),
		)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.service.GetStream(ctx, in)
}

func (s *StreamServiceWrapper) UpdateStream(ctx context.Context, in *v1.UpdateStreamRequest) (res *v1.UpdateStreamResponse, err error) {
	ctx, span := otel.Tracer(StreamServiceInstance).Start(ctx, "StreamService.UpdateStream")
	defer func() {
		span.SetAttributes(
			attribute.Stringer("in", in),
			attribute.Stringer("res", res),
		)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("event", "error"),
				attribute.String("message", err.Error()),
			)
		}
		span.End()
	}()
	return s.service.UpdateStream(ctx, in)
}
