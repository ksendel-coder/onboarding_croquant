CREATE TABLE IF NOT EXISTS tour_hint_skipped_archive (
    LIKE tour_events INCLUDING INDEXES
);

WITH skipped AS (
    DELETE FROM tour_events WHERE event_type = 'hint_skipped' RETURNING *
)
INSERT INTO tour_hint_skipped_archive SELECT * FROM skipped
ON CONFLICT DO NOTHING;

ALTER TABLE tour_events
    DROP CONSTRAINT tour_events_event_type_check,
    DROP CONSTRAINT tour_events_check,
    ADD CONSTRAINT tour_events_event_type_check CHECK (
        event_type IN ('tour_started', 'hint_shown', 'hint_completed', 'tour_completed', 'tour_dismissed')
    ),
    ADD CONSTRAINT tour_events_check CHECK (
        (event_type IN ('hint_shown', 'hint_completed') AND hint_id IS NOT NULL)
        OR (event_type IN ('tour_started', 'tour_completed', 'tour_dismissed') AND hint_id IS NULL)
    );
