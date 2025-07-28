-- Remove the CHECK constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_receiver_or_group_check;

-- Remove the conditional indexes
DROP INDEX IF EXISTS idx_messages_receiver_id;
DROP INDEX IF EXISTS idx_messages_group_id;
