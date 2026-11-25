ALTER TABLE conversations ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS class_id UUID REFERENCES classes(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_conversations_owner ON conversations(owner_id);
CREATE INDEX IF NOT EXISTS idx_conversations_class ON conversations(class_id);