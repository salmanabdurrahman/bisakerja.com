CREATE TABLE IF NOT EXISTS notification_job_events (
  id bigserial PRIMARY KEY,
  job_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_job_events_created
  ON notification_job_events (created_at ASC, id ASC);

CREATE TABLE IF NOT EXISTS notification_delivery_tasks (
  id bigserial PRIMARY KEY,
  notification_id text NOT NULL,
  user_id text NOT NULL,
  user_email text NOT NULL,
  user_name text NOT NULL,
  job_id text NOT NULL,
  channel text NOT NULL,
  job_title text NOT NULL,
  company text NOT NULL DEFAULT '',
  location text NOT NULL DEFAULT '',
  url text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notification_delivery_tasks_created
  ON notification_delivery_tasks (created_at ASC, id ASC);
