INSERT INTO subject_teachers (subject_id, teacher_id)
SELECT DISTINCT subject_id, teacher_id
FROM class_subjects
WHERE teacher_id IS NOT NULL
ON CONFLICT (subject_id, teacher_id) DO NOTHING;

ALTER TABLE class_subjects
    DROP CONSTRAINT IF EXISTS class_subjects_subject_id_teacher_id_fkey;

ALTER TABLE class_subjects
    ADD CONSTRAINT class_subjects_subject_id_teacher_id_fkey
    FOREIGN KEY (subject_id, teacher_id)
    REFERENCES subject_teachers (subject_id, teacher_id)
    DEFERRABLE INITIALLY IMMEDIATE;