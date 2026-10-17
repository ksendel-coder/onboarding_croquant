CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE tours (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id      UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    enabled     BOOLEAN     NOT NULL DEFAULT false,
    priority    INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ
);

CREATE INDEX tours_resolve_idx
    ON tours (app_id, priority DESC, created_at ASC)
    WHERE enabled AND archived_at IS NULL;

CREATE TRIGGER tours_set_updated_at
    BEFORE UPDATE ON tours
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tour_versions (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id             UUID        NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
    version             INT         NOT NULL,
    status              TEXT        NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    trigger_type        TEXT        NOT NULL DEFAULT 'on_load',
    target_path         TEXT        NOT NULL,
    audience_show_once  BOOLEAN     NOT NULL DEFAULT false,
    audience_max_shows  INT         NOT NULL DEFAULT 0,
    audience_only_new   BOOLEAN     NOT NULL DEFAULT false,
    trigger_config      JSONB       DEFAULT '{}'::jsonb,
    audience_rules      JSONB       DEFAULT '[]'::jsonb,
    created_by          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at        TIMESTAMPTZ,
    archived_at         TIMESTAMPTZ,
    UNIQUE (tour_id, version),
    CHECK (status <> 'published' OR published_at IS NOT NULL),
    CHECK (status <> 'archived' OR archived_at IS NOT NULL)
);

CREATE UNIQUE INDEX tour_versions_one_published
    ON tour_versions (tour_id)
    WHERE status = 'published';

CREATE UNIQUE INDEX tour_versions_one_draft
    ON tour_versions (tour_id)
    WHERE status = 'draft';

CREATE INDEX tour_versions_history_idx
    ON tour_versions (tour_id, version DESC);