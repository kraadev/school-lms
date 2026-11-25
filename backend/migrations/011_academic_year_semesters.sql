CREATE TABLE semesters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    number SMALLINT NOT NULL CHECK (number > 0),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_academic_year_semester_number UNIQUE (academic_year_id, number),
    CONSTRAINT chk_academic_year_semester_dates CHECK (end_date > start_date)
);

INSERT INTO semesters (academic_year_id, number, start_date, end_date)
SELECT id, semester, start_date, end_date
FROM academic_years
WHERE semester IS NOT NULL AND start_date IS NOT NULL AND end_date IS NOT NULL
  AND end_date > start_date
ON CONFLICT (academic_year_id, number) DO NOTHING;

ALTER TABLE academic_years
    ALTER COLUMN semester DROP NOT NULL,
    ALTER COLUMN start_date DROP NOT NULL,
    ALTER COLUMN end_date DROP NOT NULL;