ALTER TABLE conversations ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS assigned_admin_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS closed_at TIMESTAMP NULL;

ALTER TABLE conversations DROP CONSTRAINT IF EXISTS chk_conversations_type;
ALTER TABLE conversations ADD CONSTRAINT chk_conversations_type
    CHECK (type IN ('private', 'group', 'class', 'customer_service'));

ALTER TABLE conversations ADD CONSTRAINT chk_customer_service_status
    CHECK (status IN ('waiting', 'active', 'closed'));

CREATE INDEX IF NOT EXISTS idx_conversations_cs_queue
    ON conversations(type, status, updated_at DESC);

CREATE TABLE IF NOT EXISTS faqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question VARCHAR(500) NOT NULL,
    answer TEXT NOT NULL,
    category VARCHAR(80) NOT NULL DEFAULT 'Umum',
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_faqs_published_category ON faqs(published, category);