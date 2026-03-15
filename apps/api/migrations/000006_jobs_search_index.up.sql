-- Add indexes to optimize LIKE queries on jobs table
CREATE INDEX IF NOT EXISTS idx_jobs_title_lower ON jobs (lower(title) text_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_jobs_description_lower ON jobs (lower(description) text_pattern_ops);
