package interfaces

import (
	"context"
	v1 "stream-server/api/v1"
)

type IUsecase interface {
	ListStreams(context.Context, *v1.ListStreamsRequest) ([]*v1.Stream, error)
	GetStream(ctx context.Context, in *v1.GetStreamRequest) (res *v1.Stream, err error)
	UpdateStream(ctx context.Context, in *v1.UpdateStreamRequest) (res *v1.Stream, err error)
}
