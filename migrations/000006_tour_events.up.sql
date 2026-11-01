CREATE TABLE tour_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE RESTRICT,
    event_id UUID NOT NULL,
    attempt_id UUID NOT NULL,
    subject_id TEXT NOT NULL CHECK (length(btrim(subject_id)) > 0),
    tour_id UUID NOT NULL REFERENCES tours(id) ON DELETE RESTRICT,
    tour_version_id UUID NOT NULL REFERENCES tour_versions(id) ON DELETE RESTRICT,
    hint_id UUID REFERENCES hints(id) ON DELETE RESTRICT,
    event_type TEXT NOT NULL CHECK (event_type IN (
        'tour_started', 'hint_shown', 'hint_completed', 'tour_completed', 'tour_dismissed'
    )),
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(app_id, event_id),
    CHECK (
        (event_type IN ('hint_shown', 'hint_completed') AND hint_id IS NOT NULL)
        OR (event_type IN ('tour_started', 'tour_completed', 'tour_dismissed') AND hint_id IS NULL)
    )
);

CREATE INDEX tour_events_funnel_idx ON tour_events(tour_id, tour_version_id, event_type, attempt_id);
CREATE INDEX tour_events_hint_idx ON tour_events(tour_version_id, hint_id, event_type, attempt_id);
CREATE INDEX tour_events_subject_idx ON tour_events(app_id, subject_id, tour_id, received_at);
