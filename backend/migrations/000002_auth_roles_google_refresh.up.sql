ALTER TABLE users
    ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) NOT NULL DEFAULT 'email',
    ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255),
    ALTER COLUMN password DROP NOT NULL,
    ALTER COLUMN password SET DEFAULT '',
    ALTER COLUMN role SET DEFAULT 'customer';

-- Existing 'user' role becomes 'customer'.
UPDATE users SET role = 'customer' WHERE role = 'user';

ALTER TABLE users
    ADD CONSTRAINT uni_users_google_sub UNIQUE (google_sub);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(128) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    device_info TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
