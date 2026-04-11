-- Prompt history for Telegram users
CREATE TABLE IF NOT EXISTS prompt_history (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_prompt_history_profile_created_at
    ON prompt_history (profile_id, created_at DESC);
