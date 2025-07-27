CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    receiver_id INTEGER REFERENCES users(id),
    group_id INTEGER REFERENCES groups(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages (receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_group_id ON messages (group_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);
CREATE INDEX IF NOT EXISTS idx_group_created_at ON messages (group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_receiver_created_at ON messages (receiver_id, created_at DESC);
