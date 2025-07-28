-- Add CHECK constraint to ensure either receiver_id OR group_id is set, but not both
ALTER TABLE messages ADD CONSTRAINT messages_receiver_or_group_check 
    CHECK (
        (receiver_id IS NOT NULL AND group_id IS NULL) OR 
        (receiver_id IS NULL AND group_id IS NOT NULL)
    );

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages (receiver_id) WHERE receiver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_group_id ON messages (group_id) WHERE group_id IS NOT NULL;
