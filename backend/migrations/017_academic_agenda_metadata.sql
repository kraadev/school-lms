ALTER TABLE academic_agendas ADD COLUMN IF NOT EXISTS start_time VARCHAR(5);
ALTER TABLE academic_agendas ADD COLUMN IF NOT EXISTS end_time VARCHAR(5);
ALTER TABLE academic_agendas ADD COLUMN IF NOT EXISTS status VARCHAR(40) NOT NULL DEFAULT 'Terjadwal';
ALTER TABLE academic_agendas ADD COLUMN IF NOT EXISTS created_by_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_academic_agendas_created_by ON academic_agendas(created_by_id);