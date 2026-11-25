CREATE TABLE IF NOT EXISTS quiz_attempt_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    attempt_id UUID NOT NULL,
    question_id UUID NOT NULL,
    option_id UUID NULL,

    essay_answer TEXT NULL,

    answered_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attempt_answers_attempt
        FOREIGN KEY (attempt_id)
        REFERENCES quiz_attempts(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_attempt_answers_question
        FOREIGN KEY (question_id)
        REFERENCES quiz_questions(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_attempt_answers_option
        FOREIGN KEY (option_id)
        REFERENCES quiz_question_options(id)
        ON DELETE SET NULL,

    CONSTRAINT uq_attempt_question
        UNIQUE (attempt_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_attempt_answers_attempt
    ON quiz_attempt_answers(attempt_id);

CREATE INDEX IF NOT EXISTS idx_attempt_answers_question
    ON quiz_attempt_answers(question_id);