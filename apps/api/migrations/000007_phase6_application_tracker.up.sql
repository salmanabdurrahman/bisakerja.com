-- phase6: application tracker (bookmarks + tracked applications)

CREATE TABLE IF NOT EXISTS bookmarks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, job_id)
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks (user_id);

CREATE TABLE IF NOT EXISTS tracked_applications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id      TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'applied',
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, job_id)
);

CREATE INDEX IF NOT EXISTS idx_tracked_applications_user_id ON tracked_applications (user_id);
