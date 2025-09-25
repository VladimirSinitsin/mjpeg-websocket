-- +goose NO TRANSACTION
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    frame_interval_ms INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS frames (
    id UUID PRIMARY KEY,
    stream_id UUID NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    payload BYTEA NOT NULL,
    mime_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (stream_id, sequence)
);

CREATE INDEX IF NOT EXISTS idx_frames_stream_sequence ON frames(stream_id, sequence);
-- +goose Down
