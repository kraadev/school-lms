CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL,

    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    reference_type VARCHAR(50) NULL,
    reference_id UUID NULL,

    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_notifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id
    ON notifications(user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_is_read
    ON notifications(is_read);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at
    ON notifications(created_at);

CREATE INDEX IF NOT EXISTS idx_notifications_reference
    ON notifications(reference_type, reference_id);