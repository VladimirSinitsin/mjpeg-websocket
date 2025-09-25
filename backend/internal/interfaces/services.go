package interfaces

import (
	"context"
	"net/http"
	v1 "stream-server/api/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

type IHealthService interface {
	Live(context.Context, *emptypb.Empty) (*v1.HealthReply, error)
	Ready(context.Context, *emptypb.Empty) (*v1.HealthReply, error)
}

type IStreamService interface {
	ListStreams(context.Context, *v1.ListStreamsRequest) (*v1.ListStreamsResponse, error)
	GetStream(context.Context, *v1.GetStreamRequest) (*v1.GetStreamResponse, error)
	UpdateStream(context.Context, *v1.UpdateStreamRequest) (*v1.UpdateStreamResponse, error)

	// Websocket handlers
	StreamWSHandler() http.HandlerFunc
}
