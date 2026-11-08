CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    apps_name TEXT NOT NULL,
    idol_name VARCHAR(100) NOT NULL DEFAULT '',
    provider VARCHAR(32) NOT NULL DEFAULT '',
    tokens_in INTEGER NOT NULL DEFAULT 0,
    tokens_out INTEGER NOT NULL DEFAULT 0,
    ok BOOLEAN NOT NULL DEFAULT FALSE,
    occurred_at TIMESTAMPTZ NOT NULL,
    usage_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_chat_device_date ON chat_messages (device_id, usage_date);
CREATE INDEX IF NOT EXISTS idx_chat_app_date ON chat_messages (apps_name, usage_date);
CREATE INDEX IF NOT EXISTS idx_chat_occurred ON chat_messages (occurred_at);
