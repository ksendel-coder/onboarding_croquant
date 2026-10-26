ALTER TABLE tour_progress
ADD CONSTRAINT tour_progress_current_hint_fk
FOREIGN KEY (current_hint_id)
REFERENCES hints(id)
ON DELETE RESTRICT;
