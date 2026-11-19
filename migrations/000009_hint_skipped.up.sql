ALTER TABLE tour_events
    DROP CONSTRAINT tour_events_event_type_check,
    DROP CONSTRAINT tour_events_check,
    ADD CONSTRAINT tour_events_event_type_check CHECK (
        event_type IN ('tour_started', 'hint_shown', 'hint_completed', 'hint_skipped', 'tour_completed', 'tour_dismissed')
    ),
    ADD CONSTRAINT tour_events_check CHECK (
        (event_type IN ('hint_shown', 'hint_completed', 'hint_skipped') AND hint_id IS NOT NULL)
        OR (event_type IN ('tour_started', 'tour_completed', 'tour_dismissed') AND hint_id IS NULL)
    );

DO $$
BEGIN
    IF to_regclass('tour_hint_skipped_archive') IS NOT NULL THEN
        INSERT INTO tour_events
        SELECT a.* FROM tour_hint_skipped_archive a
        JOIN apps ON apps.id = a.app_id
        JOIN tours ON tours.id = a.tour_id
        JOIN tour_versions ON tour_versions.id = a.tour_version_id
        JOIN hints ON hints.id = a.hint_id
        ON CONFLICT DO NOTHING;

        DELETE FROM tour_hint_skipped_archive a
        USING tour_events e
        WHERE e.id = a.id;
    END IF;
END $$;
