DROP INDEX IF EXISTS api_key_usage_api_key_idx;

ALTER TABLE IF EXISTS api_key_usage
    ADD COLUMN IF NOT EXISTS creation_time TIMESTAMP  DEFAULT CURRENT_TIMESTAMP(0);

CREATE INDEX IF NOT EXISTS api_key_usage_api_key_creation_time_idx 
    ON api_key_usage (api_key, creation_time);