ALTER TABLE IF EXISTS daily_app_usages
    ALTER COLUMN country TYPE VARCHAR(100);

ALTER TABLE IF EXISTS daily_app_usages
    ADD COLUMN IF NOT EXISTS country_code VARCHAR(10);

CREATE INDEX IF NOT EXISTS idx_usage_country_code ON daily_app_usages(country_code);
