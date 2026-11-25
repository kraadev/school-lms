ALTER TABLE meeting_participants
    DROP CONSTRAINT IF EXISTS fk_meetings_participants;

ALTER TABLE meeting_participants
    ADD CONSTRAINT fk_meetings_participants
    FOREIGN KEY (meeting_id) REFERENCES meetings(id) ON DELETE CASCADE;