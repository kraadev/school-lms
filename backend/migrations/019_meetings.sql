CREATE TABLE IF NOT EXISTS meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_subject_id UUID NOT NULL REFERENCES class_subjects(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type VARCHAR(20) NOT NULL DEFAULT 'meeting' CHECK (type IN ('voice', 'video', 'meeting')),
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'live', 'ended')),
    approval_required BOOLEAN NOT NULL DEFAULT TRUE,
    max_participants INTEGER NOT NULL DEFAULT 50 CHECK (max_participants > 0 AND max_participants <= 500),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meeting_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'requested' CHECK (status IN ('requested', 'approved', 'rejected', 'kicked')),
    mic_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    camera_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    joined_at TIMESTAMPTZ,
    left_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (meeting_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_meetings_host_status ON meetings(host_id, status);
CREATE INDEX IF NOT EXISTS idx_meetings_class_subject ON meetings(class_subject_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_user_status ON meeting_participants(user_id, status);
