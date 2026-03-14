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

CREATE INDEX IF NOT EXISTS idx_jobs_company_name_trgm
  ON jobs USING gin (company_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_jobs_published_at
  ON jobs (published_at DESC);

CREATE INDEX IF NOT EXISTS idx_jobs_created_at
  ON jobs (created_at DESC);
