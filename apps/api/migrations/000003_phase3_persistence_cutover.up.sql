ALTER TABLE IF EXISTS user_preferences
  ADD COLUMN IF NOT EXISTS alert_mode text NOT NULL DEFAULT 'instant',
  ADD COLUMN IF NOT EXISTS digest_hour integer,
  ALTER COLUMN updated_at DROP NOT NULL,
  ALTER COLUMN updated_at DROP DEFAULT;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'company_name'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'company'
  ) THEN
    ALTER TABLE jobs RENAME COLUMN company_name TO company;
  END IF;
END
$$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'published_at'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'posted_at'
  ) THEN
    ALTER TABLE jobs RENAME COLUMN published_at TO posted_at;
  END IF;
END
$$;

ALTER TABLE IF EXISTS jobs
  ADD COLUMN IF NOT EXISTS salary_range text,
  ADD COLUMN IF NOT EXISTS raw_data jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE IF EXISTS jobs
  DROP COLUMN IF EXISTS job_type;

ALTER TABLE IF EXISTS transactions
  ADD COLUMN IF NOT EXISTS plan_code text NOT NULL DEFAULT 'pro_monthly',
  ADD COLUMN IF NOT EXISTS checkout_url text,
  ADD COLUMN IF NOT EXISTS idempotency_key text,
  ADD COLUMN IF NOT EXISTS expires_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_transactions_user_idempotency
  ON transactions (user_id, idempotency_key);

CREATE INDEX IF NOT EXISTS idx_transactions_mayar_transaction_id
  ON transactions (mayar_transaction_id);

UPDATE transactions
SET mayar_transaction_id = NULL
WHERE mayar_transaction_id IS NOT NULL
  AND btrim(mayar_transaction_id) = '';

WITH ranked_transactions AS (
  SELECT id,
         row_number() OVER (
           PARTITION BY provider, mayar_transaction_id
           ORDER BY updated_at DESC, created_at DESC, id DESC
         ) AS row_number
  FROM transactions
  WHERE mayar_transaction_id IS NOT NULL
)
UPDATE transactions AS t
SET mayar_transaction_id = NULL
FROM ranked_transactions AS ranked
WHERE t.id = ranked.id
  AND ranked.row_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_provider_mayar_transaction
  ON transactions (provider, mayar_transaction_id)
  WHERE mayar_transaction_id IS NOT NULL;

ALTER TABLE IF EXISTS notifications
  ADD COLUMN IF NOT EXISTS error_message text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS read_at timestamptz;

ALTER TABLE IF EXISTS webhook_deliveries
  ADD COLUMN IF NOT EXISTS error_message text NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS saved_searches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  query text NOT NULL,
  location text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT '',
  salary_min bigint,
  frequency text NOT NULL CHECK (frequency IN ('instant', 'daily_digest', 'weekly_digest')),
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_searches_dedupe
  ON saved_searches (user_id, lower(query), lower(location), lower(source), salary_min, frequency);

CREATE INDEX IF NOT EXISTS idx_saved_searches_user_created
  ON saved_searches (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS watchlist_companies (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  company_slug text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, company_slug)
);

CREATE INDEX IF NOT EXISTS idx_watchlist_companies_user_created
  ON watchlist_companies (user_id, created_at DESC);
