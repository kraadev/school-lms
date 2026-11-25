CREATE TABLE IF NOT EXISTS academic_agendas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    title VARCHAR(180) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    category VARCHAR(80) NOT NULL DEFAULT 'Kegiatan sekolah',
    audience VARCHAR(40) NOT NULL DEFAULT 'Semua',
    location VARCHAR(180),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_academic_agenda_dates CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_academic_agendas_year_dates
ON academic_agendas(academic_year_id, start_date, end_date);