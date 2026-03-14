DROP INDEX IF EXISTS idx_watchlist_companies_user_created;
DROP TABLE IF EXISTS watchlist_companies;

DROP INDEX IF EXISTS idx_saved_searches_user_created;
DROP INDEX IF EXISTS idx_saved_searches_dedupe;
DROP TABLE IF EXISTS saved_searches;

ALTER TABLE IF EXISTS webhook_deliveries
  DROP COLUMN IF EXISTS error_message;

ALTER TABLE IF EXISTS notifications
  DROP COLUMN IF EXISTS read_at,
  DROP COLUMN IF EXISTS error_message;

DROP INDEX IF EXISTS idx_transactions_mayar_transaction_id;
DROP INDEX IF EXISTS idx_transactions_user_idempotency;
DROP INDEX IF EXISTS idx_transactions_provider_mayar_transaction;

ALTER TABLE IF EXISTS transactions
  DROP COLUMN IF EXISTS expires_at,
  DROP COLUMN IF EXISTS idempotency_key,
  DROP COLUMN IF EXISTS checkout_url,
  DROP COLUMN IF EXISTS plan_code;

ALTER TABLE IF EXISTS jobs
  ADD COLUMN IF NOT EXISTS job_type text;

ALTER TABLE IF EXISTS jobs
  DROP COLUMN IF EXISTS raw_data,
  DROP COLUMN IF EXISTS salary_range;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'posted_at'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'published_at'
  ) THEN
    ALTER TABLE jobs RENAME COLUMN posted_at TO published_at;
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
      AND column_name = 'company'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'company_name'
  ) THEN
    ALTER TABLE jobs RENAME COLUMN company TO company_name;
  END IF;
END
$$;

ALTER TABLE IF EXISTS user_preferences
  DROP COLUMN IF EXISTS digest_hour,
  DROP COLUMN IF EXISTS alert_mode;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'user_preferences'
      AND column_name = 'updated_at'
  ) THEN
    UPDATE user_preferences
    SET updated_at = now()
    WHERE updated_at IS NULL;

    ALTER TABLE user_preferences
      ALTER COLUMN updated_at SET DEFAULT now(),
      ALTER COLUMN updated_at SET NOT NULL;
  END IF;
END
$$;
