package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "stream-server/api/v1"
	dbrepo "stream-server/internal/data/repo"
	"stream-server/internal/interfaces"

	conf "stream-server/config"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubRepo struct {
	rows []dbrepo.ListStreamsRow
	err  error
}

func (s *stubRepo) ListStreams(_ context.Context) ([]dbrepo.ListStreamsRow, error) {
	return s.rows, s.err
}

func TestStreamUsecase_ListStreams_Success(t *testing.T) {
	now := time.Unix(1700000001, 0).UTC()
	uuid := pgtype.UUID{}
	_ = uuid.Scan("84a1c6a6-96ee-4d7b-94a9-0f3fbb29e7a1")
	repo := &stubRepo{
		rows: []dbrepo.ListStreamsRow{{
			ID:              uuid,
			Title:           "t",
			Description:     "d",
			FrameIntervalMs: 40,
			CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
			FrameCount:      7,
		}},
	}

	uc := NewStreamUsecase(repo, log.NewHelper(log.NewStdLogger(nil)), &conf.Config{})
	got, err := uc.ListStreams(context.Background(), &v1.ListStreamsRequest{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got == nil || len(got) != 1 {
		t.Fatalf("unexpected result: %#v", got)
	}
	if got[0].Title != "t" || got[0].FrameCount != 7 {
		t.Fatalf("unexpected mapping: %#v", got[0])
	}
}

func TestStreamUsecase_ListStreams_RepoError(t *testing.T) {
	want := errors.New("db err")
	uc := NewStreamUsecase(&stubRepo{err: want}, log.NewHelper(log.NewStdLogger(nil)), &conf.Config{})
	got, err := uc.ListStreams(context.Background(), &v1.ListStreamsRequest{})
	if err == nil {
		t.Fatalf("expected error, got %#v", got)
	}
	if !errors.Is(err, want) {
		t.Fatalf("expected Is(%v), got %v", want, err)
	}
}

var _ interfaces.IRepo = (*stubRepo)(nil)
