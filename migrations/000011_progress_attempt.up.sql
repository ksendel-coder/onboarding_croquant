ALTER TABLE tour_progress ADD COLUMN attempt_id UUID;

UPDATE tour_progress p
SET attempt_id = (
    SELECT e.attempt_id
    FROM tour_events e
    WHERE e.app_id = p.app_id
      AND e.subject_id = p.subject_id
      AND e.tour_id = p.tour_id
      AND e.tour_version_id = p.tour_version_id
      AND e.received_at >= p.started_at
    ORDER BY e.received_at DESC, e.id DESC
    LIMIT 1
);
