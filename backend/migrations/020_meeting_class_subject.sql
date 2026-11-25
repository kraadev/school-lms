ALTER TABLE meetings ADD COLUMN IF NOT EXISTS class_subject_id UUID;
ALTER TABLE meetings DROP CONSTRAINT IF EXISTS meetings_class_subject_id_fkey;
ALTER TABLE meetings ADD CONSTRAINT meetings_class_subject_id_fkey
    FOREIGN KEY (class_subject_id) REFERENCES class_subjects(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_meetings_class_subject ON meetings(class_subject_id);

-- Existing meetings must be assigned to a class subject before this column can be mandatory.
-- The application rejects new meetings without class_subject_id.