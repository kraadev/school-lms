CREATE TABLE IF NOT EXISTS subject_teachers (
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (subject_id, teacher_id)
);

CREATE INDEX IF NOT EXISTS idx_subject_teachers_teacher ON subject_teachers(teacher_id);