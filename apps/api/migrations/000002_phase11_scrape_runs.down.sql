DROP INDEX IF EXISTS idx_jobs_created_at;
DROP INDEX IF EXISTS idx_jobs_published_at;
DROP INDEX IF EXISTS idx_jobs_company_name_trgm;
DROP INDEX IF EXISTS idx_jobs_title_trgm;
DROP INDEX IF EXISTS idx_scrape_runs_source_created;
DROP TABLE IF EXISTS scrape_runs;
