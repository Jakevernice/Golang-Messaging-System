CREATE TABLE IF NOT EXISTS group_members (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id),
    member_id INTEGER NOT NULL REFERENCES users(id),
    is_admin BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (group_id, member_id)
);

-- Index to speed up processes in a group
CREATE INDEX IF NOT EXISTS idx_group_members_group_id ON group_members (group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_pair ON group_members (group_id, member_id);
CREATE INDEX IF NOT EXISTS idx_group_admin_check ON group_members (group_id, is_admin);
