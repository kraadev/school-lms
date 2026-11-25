CREATE TABLE majors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO majors (code, name)
SELECT DISTINCT lower(regexp_replace(trim(major), '[^a-zA-Z0-9]+', '-', 'g')), trim(major)
FROM classes
WHERE major IS NOT NULL AND trim(major) <> ''
ON CONFLICT (code) DO NOTHING;

ALTER TABLE classes ADD COLUMN major_id UUID REFERENCES majors(id) ON DELETE SET NULL;

UPDATE classes c
SET major_id = m.id
FROM majors m
WHERE c.major IS NOT NULL
  AND lower(regexp_replace(trim(c.major), '[^a-zA-Z0-9]+', '-', 'g')) = m.code;