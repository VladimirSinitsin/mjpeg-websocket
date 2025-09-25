package service

import (
	"context"
	"testing"

	v1 "stream-server/api/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestHealthService_Live_OK(t *testing.T) {
	svc := NewHealthService(nil)
	got, err := svc.Live(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got == nil || got.Status != "ok" {
		t.Fatalf("unexpected resp: %#v", got)
	}
}

func TestHealthService_Ready_DBNil_Error(t *testing.T) {
	svc := NewHealthService(nil)
	resp, err := svc.Ready(context.Background(), &emptypb.Empty{})
	if err == nil || resp != nil {
		t.Fatalf("expected error when db is nil, got resp=%#v err=%v", resp, err)
	}
}

var _ = v1.HealthReply{}
