CREATE TABLE IF NOT EXISTS ai_usage_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  feature text NOT NULL CHECK (feature IN (
    'search_assistant',
    'job_fit_summary',
    'cover_letter_draft',
    'interview_prep'
  )),
  tier text NOT NULL CHECK (tier IN ('free', 'premium')),
  provider text NOT NULL DEFAULT '',
  model text NOT NULL DEFAULT '',
  tokens_in integer NOT NULL DEFAULT 0 CHECK (tokens_in >= 0),
  tokens_out integer NOT NULL DEFAULT 0 CHECK (tokens_out >= 0),
  prompt_hash text NOT NULL DEFAULT '',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_user_feature_created
  ON ai_usage_logs (user_id, feature, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_created
  ON ai_usage_logs (created_at DESC);
