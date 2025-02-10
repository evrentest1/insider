    BEGIN TRANSACTION;

    CREATE TYPE message_sending_status AS ENUM ('pending', 'in-progress', 'success', 'failed');

    CREATE TABLE IF NOT EXISTS messages
    (
        id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        content      CHARACTER VARYING(160) NOT NULL,
        phone_number TEXT                   NOT NULL,
        status       message_sending_status NOT NULL DEFAULT 'pending'::message_sending_status,
        message_id   TEXT,
        created_at   TIMESTAMP WITH TIME ZONE        DEFAULT NOW(),
        updated_at   TIMESTAMP WITH TIME ZONE
    );

    CREATE INDEX IF NOT EXISTS idx_messages_status ON messages USING btree (status, created_at);

    COMMIT;
