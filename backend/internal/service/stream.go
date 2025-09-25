package service

import (
	"context"
	"net/http"
	"stream-server/internal/biz/session/store_pool"

	"github.com/go-kratos/kratos/v2/log"

	v1 "stream-server/api/v1"
	"stream-server/internal/interfaces"
)

type StreamService struct {
	v1.UnimplementedStreamServiceServer

	uc    interfaces.IUsecase
	log   *log.Helper
	store *store_pool.ChunkStore
}

func NewStreamService(uc interfaces.IUsecase, l *log.Helper, store *store_pool.ChunkStore) *StreamService {
	return &StreamService{
		uc:    uc,
		log:   l,
		store: store,
	}
}

func (s *StreamService) ListStreams(ctx context.Context, in *v1.ListStreamsRequest) (res *v1.ListStreamsResponse, err error) {
	streams, err := s.uc.ListStreams(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.ListStreamsResponse{
		Streams: streams,
	}, err
}

func (s *StreamService) GetStream(ctx context.Context, in *v1.GetStreamRequest) (res *v1.GetStreamResponse, err error) {
	stream, err := s.uc.GetStream(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.GetStreamResponse{
		Stream: stream,
	}, err
}

func (s *StreamService) UpdateStream(ctx context.Context, in *v1.UpdateStreamRequest) (res *v1.UpdateStreamResponse, err error) {
	stream, err := s.uc.UpdateStream(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateStreamResponse{
		Stream: stream,
	}, err
}

func (s *StreamService) StreamWSHandler() http.HandlerFunc {
	return WSStreamHandler(s.store)
}
