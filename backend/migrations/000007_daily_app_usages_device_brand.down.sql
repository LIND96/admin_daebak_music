DROP INDEX IF EXISTS idx_usage_brand;
ALTER TABLE IF EXISTS daily_app_usages
    DROP COLUMN IF EXISTS device_model,
    DROP COLUMN IF EXISTS device_brand;
