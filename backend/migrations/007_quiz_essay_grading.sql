ALTER TABLE quiz_attempt_answers
ADD COLUMN IF NOT EXISTS score DOUBLE PRECISION NULL;

ALTER TABLE quiz_attempt_answers
ADD COLUMN IF NOT EXISTS feedback TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_attempt_answers_score
ON quiz_attempt_answers(score);