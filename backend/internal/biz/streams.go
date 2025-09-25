package biz

import (
	"context"
	"fmt"

	v1 "stream-server/api/v1"
	"stream-server/internal/converters"
)

// ListStreams gets streams
func (u *StreamUsecase) ListStreams(ctx context.Context, _ *v1.ListStreamsRequest) (_ []*v1.Stream, err error) {
	streamRows, err := u.repo.ListStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get streams: %w", err)
	}

	return converters.ToApiStreamResponseList(streamRows), nil
}

// GetStream get stream by ID
func (u *StreamUsecase) GetStream(ctx context.Context, in *v1.GetStreamRequest) (res *v1.Stream, err error) {
	uuid, err := converters.StringToPgUUID(in.Id)
	if err != nil {
		return nil, fmt.Errorf("error converting uuid: %w", err)
	}

	stream, err := u.repo.GetStream(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("error get stream: %w", err)
	}

	return converters.ToApiStreamResponse(stream), nil
}

// UpdateStream update stream
func (u *StreamUsecase) UpdateStream(ctx context.Context, in *v1.UpdateStreamRequest) (res *v1.Stream, err error) {
	params, err := converters.ToDbUpdateStreamParams(in)
	if err != nil {
		return nil, fmt.Errorf("error converting params: %w", err)
	}

	stream, err := u.repo.UpdateStream(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error update stream: %w", err)
	}

	return converters.ToApiStreamUpdateResult(stream), nil
}
