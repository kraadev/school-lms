CREATE TABLE IF NOT EXISTS assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_subject_id UUID NOT NULL,
    teacher_id UUID NOT NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT,

    deadline TIMESTAMP NOT NULL,

    max_score INTEGER NOT NULL DEFAULT 100,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    published_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_assignments_class_subject
        FOREIGN KEY (class_subject_id)
        REFERENCES class_subjects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_assignments_teacher
        FOREIGN KEY (teacher_id)
        REFERENCES teachers(id)
        ON DELETE CASCADE,

    CONSTRAINT assignments_status_check
        CHECK (status IN ('draft', 'published')),

    CONSTRAINT assignments_max_score_check
        CHECK (max_score > 0)
);

CREATE INDEX IF NOT EXISTS idx_assignments_class_subject_id
ON assignments(class_subject_id);

CREATE INDEX IF NOT EXISTS idx_assignments_teacher_id
ON assignments(teacher_id);

CREATE INDEX IF NOT EXISTS idx_assignments_deadline
ON assignments(deadline);

CREATE INDEX IF NOT EXISTS idx_assignments_status
ON assignments(status);

CREATE TABLE IF NOT EXISTS assignment_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    assignment_id UUID NOT NULL,

    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_assignment_files_assignment
        FOREIGN KEY (assignment_id)
        REFERENCES assignments(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS assignment_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    assignment_id UUID NOT NULL,
    student_id UUID NOT NULL,

    submitted_at TIMESTAMP,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    score NUMERIC(5,2),
    feedback TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_assignment_submissions_assignment
        FOREIGN KEY (assignment_id)
        REFERENCES assignments(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_assignment_submissions_student
        FOREIGN KEY (student_id)
        REFERENCES students(id)
        ON DELETE CASCADE,

    CONSTRAINT assignment_submissions_status_check
        CHECK (status IN ('draft', 'submitted', 'late', 'graded')),

    CONSTRAINT assignment_submissions_score_check
        CHECK (score IS NULL OR score >= 0),

    CONSTRAINT unique_assignment_student
        UNIQUE (assignment_id, student_id)
);

CREATE TABLE IF NOT EXISTS assignment_submission_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    submission_id UUID NOT NULL,

    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_submission_files_submission
        FOREIGN KEY (submission_id)
        REFERENCES assignment_submissions(id)
        ON DELETE CASCADE
);
