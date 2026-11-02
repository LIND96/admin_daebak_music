DROP INDEX IF EXISTS idx_usage_country_code;
ALTER TABLE IF EXISTS daily_app_usages DROP COLUMN IF EXISTS country_code;
-- country size intentionally left as VARCHAR(100) on downgrade
