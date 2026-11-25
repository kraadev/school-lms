CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    type VARCHAR(20) NOT NULL DEFAULT 'private',
    name VARCHAR(255) NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_conversations_type
        CHECK (type IN ('private', 'group', 'class'))
);

CREATE INDEX IF NOT EXISTS idx_conversations_type
    ON conversations(type);


CREATE TABLE IF NOT EXISTS conversation_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,

    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_conversation_participants_conversation
        FOREIGN KEY (conversation_id)
        REFERENCES conversations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_conversation_participants_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_conversation_participant
        UNIQUE (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_conversation_participants_conversation
    ON conversation_participants(conversation_id);

CREATE INDEX IF NOT EXISTS idx_conversation_participants_user
    ON conversation_participants(user_id);


CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL,

    message_type VARCHAR(20) NOT NULL DEFAULT 'text',
    content TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_messages_conversation
        FOREIGN KEY (conversation_id)
        REFERENCES conversations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_messages_sender
        FOREIGN KEY (sender_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_messages_type
        CHECK (message_type IN ('text', 'file', 'image', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation
    ON messages(conversation_id);

CREATE INDEX IF NOT EXISTS idx_messages_sender
    ON messages(sender_id);

CREATE INDEX IF NOT EXISTS idx_messages_created_at
    ON messages(created_at);