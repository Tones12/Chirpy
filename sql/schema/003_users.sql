-- +goose Up
ALTER TABLE users ADD COLUMN hashed_passwords TEXT NOT NULL DEFAULT 'unset';
ALTER TABLE users ALTER COLUMN hashed_passwords DROP DEFAULT;

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_passwords;