package service

import (
	"context"
	"errors"
	"testing"

	v1 "stream-server/api/v1"
	"stream-server/internal/interfaces"

	"github.com/go-kratos/kratos/v2/log"
)

type stubUsecase struct {
	resp []*v1.Stream
	err  error
}

func (s *stubUsecase) ListStreams(_ context.Context, _ *v1.ListStreamsRequest) ([]*v1.Stream, error) {
	return s.resp, s.err
}

func TestStreamService_ListStreams_Success(t *testing.T) {
	uc := &stubUsecase{
		resp: []*v1.Stream{{Id: "id-1", Title: "name"}},
	}
	svc := &StreamService{uc: uc, log: log.NewHelper(log.NewStdLogger(nil))}
	got, err := svc.ListStreams(context.Background(), &v1.ListStreamsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got.Streams) != 1 || got.Streams[0].Id != "id-1" {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestStreamService_ListStreams_Error(t *testing.T) {
	wantErr := errors.New("boom")
	uc := &stubUsecase{err: wantErr}
	svc := &StreamService{uc: uc, log: log.NewHelper(log.NewStdLogger(nil))}
	got, err := svc.ListStreams(context.Background(), &v1.ListStreamsRequest{})
	if err == nil {
		t.Fatalf("expected error, got nil and resp=%#v", got)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected Is(wantErr), got: %v", err)
	}
}

// Ensure handler is returned (not testing WS logic here)
func TestStreamService_StreamWSHandler_NotNil(t *testing.T) {
	uc := &stubUsecase{}
	svc := &StreamService{uc: uc, log: log.NewHelper(log.NewStdLogger(nil))}
	if fn := svc.StreamWSHandler(); fn == nil {
		t.Fatal("expected non-nil handler")
	}
}

// Compile-time interface check
var _ interfaces.IUsecase = (*stubUsecase)(nil)
