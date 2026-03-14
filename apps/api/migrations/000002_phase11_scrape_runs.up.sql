CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS scrape_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source text NOT NULL,
  status text NOT NULL CHECK (status IN ('success', 'partial', 'failed', 'failed_auth')),
  error_class text,
  error_message text,
  fetched_count integer NOT NULL DEFAULT 0,
  inserted_count integer NOT NULL DEFAULT 0,
  duplicate_count integer NOT NULL DEFAULT 0,
  started_at timestamptz NOT NULL,
  finished_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_scrape_runs_source_created
  ON scrape_runs (source, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_jobs_title_trgm
  ON jobs USING gin (title gin_trgm_ops);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'company_name'
  ) THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_company_name_trgm ON jobs USING gin (company_name gin_trgm_ops)';
  ELSIF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'company'
  ) THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_company_name_trgm ON jobs USING gin (company gin_trgm_ops)';
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
  ) THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_published_at ON jobs (published_at DESC)';
  ELSIF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'jobs'
      AND column_name = 'posted_at'
  ) THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_jobs_published_at ON jobs (posted_at DESC)';
  END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_jobs_created_at
  ON jobs (created_at DESC);
