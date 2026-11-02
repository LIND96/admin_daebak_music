ALTER TABLE IF EXISTS daily_app_usages
    ADD COLUMN IF NOT EXISTS device_brand VARCHAR(100),
    ADD COLUMN IF NOT EXISTS device_model VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_usage_brand ON daily_app_usages(device_brand);
