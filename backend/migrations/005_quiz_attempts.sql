CREATE TABLE IF NOT EXISTS quiz_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    quiz_id UUID NOT NULL,
    student_id UUID NOT NULL,

    attempt_number INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',

    started_at TIMESTAMP NOT NULL,
    deadline_at TIMESTAMP NOT NULL,
    submitted_at TIMESTAMP NULL,

    score DOUBLE PRECISION NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_quiz_attempts_quiz
        FOREIGN KEY (quiz_id)
        REFERENCES quizzes(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_quiz_attempts_student
        FOREIGN KEY (student_id)
        REFERENCES students(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_quiz_attempts_number
        UNIQUE (quiz_id, student_id, attempt_number),

    CONSTRAINT chk_quiz_attempts_status
        CHECK (
            status IN (
                'in_progress',
                'submitted',
                'expired',
                'graded'
            )
        ),

    CONSTRAINT chk_quiz_attempts_number
        CHECK (attempt_number > 0)
);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_quiz
    ON quiz_attempts(quiz_id);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_student
    ON quiz_attempts(student_id);

CREATE INDEX IF NOT EXISTS idx_quiz_attempts_status
    ON quiz_attempts(status);