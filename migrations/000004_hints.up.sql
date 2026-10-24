CREATE TABLE hints (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tour_version_id UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE CASCADE,
  step            INT         NOT NULL,
  title           TEXT        NOT NULL,
  content         TEXT        NOT NULL,
  selector        TEXT        NOT NULL DEFAULT '',
  placement       TEXT        NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (tour_version_id, step) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE TRIGGER hints_set_updated_at
  BEFORE UPDATE ON hints
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();