ALTER TABLE prompt_history
    ADD COLUMN IF NOT EXISTS session_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_prompt_history_profile_session
    ON prompt_history (profile_id, session_id, created_at DESC);
