DROP TABLE IF EXISTS refresh_tokens;
ALTER TABLE users DROP CONSTRAINT IF EXISTS uni_users_google_sub;
ALTER TABLE users DROP COLUMN IF EXISTS google_sub;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
