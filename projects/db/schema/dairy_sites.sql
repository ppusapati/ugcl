CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS dairy_sites (
     id         BIGSERIAL PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name_of_site TEXT NOT NULL,
    todays_work TEXT,
    employee_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    submitted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Optional: Index for employee lookups
CREATE INDEX IF NOT EXISTS idx_dairy_sites_employee_id ON dairy_sites (employee_id);

-- Optional: Index for soft-delete filtering
CREATE INDEX IF NOT EXISTS idx_dairy_sites_not_deleted ON dairy_sites (deleted_at) WHERE deleted_at IS NULL;


CREATE OR REPLACE VIEW dairy_sites_with_user AS
SELECT
    ds.id,
    ds.uuid,
    ds.name_of_site,
    ds.todays_work,
    u.username,
    u.fullname,
    ds.latitude,
    ds.longitude,
    ds.submitted_at,
    ds.created_at,
    ds.updated_at
FROM dairy_sites ds
JOIN users u ON ds.employee_id = u.id
WHERE ds.deleted_at IS NULL
  AND u.deleted_at IS NULL;
