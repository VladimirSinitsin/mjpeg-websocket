package converters

import (
	"testing"
	"time"

	v1 "stream-server/api/v1"
	dbrepo "stream-server/internal/data/repo"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestToApiStreamResponseList_ConvertsFields(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	uuid := pgtype.UUID{}
	_ = uuid.Scan("2b6f9f5e-7a7c-4d9b-8f0f-4ef8b1c4ee11")

	rows := []dbrepo.ListStreamsRow{
		{
			ID:              uuid,
			Title:           "Title A",
			Description:     "Desc A",
			FrameIntervalMs: 33,
			CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
			FrameCount:      42,
		},
	}

	got := ToApiStreamResponseList(rows)
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	item := got[0]
	if item.Id == "" {
		t.Errorf("expected non-empty Id")
	}
	if item.Title != rows[0].Title {
		t.Errorf("Title mismatch: got %q want %q", item.Title, rows[0].Title)
	}
	if item.Description != rows[0].Description {
		t.Errorf("Description mismatch: got %q want %q", item.Description, rows[0].Description)
	}
	if item.FrameIntervalMs != rows[0].FrameIntervalMs {
		t.Errorf("FrameIntervalMs mismatch: got %d want %d", item.FrameIntervalMs, rows[0].FrameIntervalMs)
	}
	if item.CreatedAt == nil || item.CreatedAt.AsTime().UTC() != now {
		t.Errorf("CreatedAt mismatch: got %v want %v", item.CreatedAt, now)
	}
	if item.UpdatedAt == nil || item.UpdatedAt.AsTime().UTC() != now {
		t.Errorf("UpdatedAt mismatch: got %v want %v", item.UpdatedAt, now)
	}
	if item.FrameCount != rows[0].FrameCount {
		t.Errorf("FrameCount mismatch: got %d want %d", item.FrameCount, rows[0].FrameCount)
	}
	// sanity check type
	var _ *v1.Stream = item
}
