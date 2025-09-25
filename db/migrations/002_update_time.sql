-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE streams ADD COLUMN IF NOT EXISTS "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- +goose Down
