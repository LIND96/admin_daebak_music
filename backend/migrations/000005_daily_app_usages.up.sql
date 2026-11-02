CREATE TABLE IF NOT EXISTS daily_app_usages (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(255) NOT NULL,
    apps_name VARCHAR(100) NOT NULL,
    android_version VARCHAR(50),
    country VARCHAR(10),
    email VARCHAR(255),
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    usage_date DATE NOT NULL DEFAULT (NOW() AT TIME ZONE 'Asia/Jakarta')::DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_app_date ON daily_app_usages(apps_name, usage_date);
CREATE INDEX IF NOT EXISTS idx_usage_device_date ON daily_app_usages(device_id, usage_date);
CREATE INDEX IF NOT EXISTS idx_usage_occurred_at ON daily_app_usages(occurred_at);
