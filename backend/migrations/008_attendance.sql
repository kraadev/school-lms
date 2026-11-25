CREATE TABLE IF NOT EXISTS attendance_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    class_subject_id UUID NOT NULL,
    teacher_id UUID NOT NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT NULL,

    attendance_date DATE NOT NULL,

    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'open',

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attendance_sessions_class_subject
        FOREIGN KEY (class_subject_id)
        REFERENCES class_subjects(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_attendance_sessions_teacher
        FOREIGN KEY (teacher_id)
        REFERENCES teachers(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_attendance_sessions_status
        CHECK (
            status IN (
                'open',
                'closed'
            )
        ),

    CONSTRAINT chk_attendance_sessions_time
        CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_class_subject
    ON attendance_sessions(class_subject_id);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_teacher
    ON attendance_sessions(teacher_id);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_date
    ON attendance_sessions(attendance_date);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_status
    ON attendance_sessions(status);

    CREATE TABLE IF NOT EXISTS attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    attendance_session_id UUID NOT NULL,
    student_id UUID NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'present',

    checked_at TIMESTAMP NULL,

    note TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_attendance_records_session
        FOREIGN KEY (attendance_session_id)
        REFERENCES attendance_sessions(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_attendance_records_student
        FOREIGN KEY (student_id)
        REFERENCES students(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_attendance_session_student
        UNIQUE (attendance_session_id, student_id),

    CONSTRAINT chk_attendance_records_status
        CHECK (
            status IN (
                'present',
                'late',
                'absent',
                'excused'
            )
        )
);

CREATE INDEX IF NOT EXISTS idx_attendance_records_session
    ON attendance_records(attendance_session_id);

CREATE INDEX IF NOT EXISTS idx_attendance_records_student
    ON attendance_records(student_id);

CREATE INDEX IF NOT EXISTS idx_attendance_records_status
    ON attendance_records(status);