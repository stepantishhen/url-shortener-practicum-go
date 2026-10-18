ALTER TABLE urls ADD COLUMN IF NOT EXISTS user_id UUID;
CREATE INDEX IF NOT EXISTS urls_user_id_idx ON urls (user_id);
