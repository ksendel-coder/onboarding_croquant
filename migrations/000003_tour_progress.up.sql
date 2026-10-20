CREATE TABLE tour_progress (
  app_id          UUID        NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
  subject_id      TEXT        NOT NULL,
  tour_id         UUID        NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
  tour_version_id UUID        NOT NULL REFERENCES tour_versions(id) ON DELETE RESTRICT,
  current_hint_id UUID,
  status          TEXT        NOT NULL,
  shows_count     INT         NOT NULL DEFAULT 0,
  started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at     TIMESTAMPTZ,

  PRIMARY KEY (app_id, subject_id, tour_id),

  CHECK (
    (finished_at IS NULL) = (status = 'in_progress')
  )
);

CREATE INDEX tour_progress_tour_status_idx
  ON tour_progress (tour_id, status);

CREATE TRIGGER tour_progress_set_updated_at
  BEFORE UPDATE ON tour_progress
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
