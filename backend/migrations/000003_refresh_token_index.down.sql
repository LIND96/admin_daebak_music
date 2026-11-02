DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
ALTER TABLE refresh_tokens ADD CONSTRAINT uni_refresh_tokens_hash UNIQUE (token_hash);
