-- GORM AutoMigrate expects its own unique-index name (idx_refresh_tokens_token_hash),
-- not a named table constraint. Align existing DBs installed via 000002.
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS uni_refresh_tokens_hash;
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
