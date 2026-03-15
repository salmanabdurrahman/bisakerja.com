-- Drop indexes created in 000006_jobs_search_index.up.sql
DROP INDEX IF EXISTS idx_jobs_title_lower;
DROP INDEX IF EXISTS idx_jobs_description_lower;
