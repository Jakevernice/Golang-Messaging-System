-- Test script to verify the new message constraint and populate test data

-- ===============================================
-- 1. CREATE TEST USERS
-- ===============================================
INSERT INTO users (username, mobile_no, password) VALUES 
('alice', '1234567890', '$2a$10$hash1'),
('bob', '1234567891', '$2a$10$hash2'), 
('charlie', '1234567892', '$2a$10$hash3'),
('diana', '1234567893', '$2a$10$hash4'),
('eve', '1234567894', '$2a$10$hash5'),
('frank', '1234567895', '$2a$10$hash6');

-- ===============================================
-- 2. CREATE TEST GROUPS
-- ===============================================
INSERT INTO groups (group_name, creator_id) VALUES 
('Development Team', 1),     -- Alice creates dev team
('Marketing Team', 2),       -- Bob creates marketing team  
('Executive Board', 3),      -- Charlie creates exec board
('Project Alpha', 1),        -- Alice creates project alpha
('Social Club', 4);          -- Diana creates social club

-- ===============================================
-- 3. ADD GROUP MEMBERS WITH ADMINS
-- ===============================================

-- Development Team (Group ID: 1)
-- Alice (creator) is auto-admin, add Bob as second admin, others as members
INSERT INTO group_members (group_id, member_id, is_admin) VALUES 
(1, 1, true),   -- Alice (creator/admin)
(1, 2, true),   -- Bob (second admin) 
(1, 3, false),  -- Charlie (member)
(1, 4, false),  -- Diana (member)
(1, 5, false);  -- Eve (member)

-- Marketing Team (Group ID: 2) 
-- Bob (creator) is auto-admin, add Diana as second admin
INSERT INTO group_members (group_id, member_id, is_admin) VALUES 
(2, 2, true),   -- Bob (creator/admin)
(2, 4, true),   -- Diana (second admin)
(2, 1, false),  -- Alice (member)
(2, 6, false);  -- Frank (member)

-- Executive Board (Group ID: 3)
-- Charlie (creator) is auto-admin, add Alice as second admin  
INSERT INTO group_members (group_id, member_id, is_admin) VALUES 
(3, 3, true),   -- Charlie (creator/admin)
(3, 1, true),   -- Alice (second admin)
(3, 2, false);  -- Bob (member)

-- Project Alpha (Group ID: 4) 
-- Alice (creator) is auto-admin, small team
INSERT INTO group_members (group_id, member_id, is_admin) VALUES 
(4, 1, true),   -- Alice (creator/admin)
(4, 3, false),  -- Charlie (member)
(4, 5, false);  -- Eve (member)

-- Social Club (Group ID: 5)
-- Diana (creator) is auto-admin, large group
INSERT INTO group_members (group_id, member_id, is_admin) VALUES 
(5, 4, true),   -- Diana (creator/admin)
(5, 1, false),  -- Alice (member)
(5, 2, false),  -- Bob (member) 
(5, 3, false),  -- Charlie (member)
(5, 5, false),  -- Eve (member)
(5, 6, false);  -- Frank (member)

-- ===============================================
-- 4. TEST DIRECT MESSAGES
-- ===============================================

-- Alice to Bob
INSERT INTO messages (sender_id, receiver_id, content) 
VALUES (1, 2, 'Hey Bob, how is the project going?');

-- Bob to Alice  
INSERT INTO messages (sender_id, receiver_id, content)
VALUES (2, 1, 'Going well! Should be done by Friday.');

-- Charlie to Diana
INSERT INTO messages (sender_id, receiver_id, content)
VALUES (3, 4, 'Can you review the marketing proposal?');

-- Diana to Charlie
INSERT INTO messages (sender_id, receiver_id, content) 
VALUES (4, 3, 'Sure, I will look at it this afternoon.');

-- Eve to Frank
INSERT INTO messages (sender_id, receiver_id, content)
VALUES (5, 6, 'Are you coming to the social club meeting?');

-- Frank to Eve
INSERT INTO messages (sender_id, receiver_id, content)
VALUES (6, 5, 'Yes, I will be there at 3 PM.');

-- ===============================================
-- 5. TEST GROUP MESSAGES  
-- ===============================================

-- Development Team messages
INSERT INTO messages (sender_id, group_id, content) VALUES 
(1, 1, 'Welcome everyone to the development team!'),
(2, 1, 'Thanks Alice! Excited to work with everyone.'),
(3, 1, 'Looking forward to collaborating on new features.'),
(4, 1, 'Happy to be part of the team!'),
(5, 1, 'Let me know how I can contribute.');

-- Marketing Team messages
INSERT INTO messages (sender_id, group_id, content) VALUES 
(2, 2, 'Our Q1 campaign is launching next week.'),
(4, 2, 'I have prepared the social media content.'),
(1, 2, 'The budget looks good from engineering perspective.'),
(6, 2, 'Website analytics are showing positive trends.');

-- Executive Board messages  
INSERT INTO messages (sender_id, group_id, content) VALUES 
(3, 3, 'Monthly board meeting scheduled for next Tuesday.'),
(1, 3, 'Technical roadmap presentation is ready.'),
(2, 3, 'Marketing metrics show 25% growth this quarter.');

-- Project Alpha messages
INSERT INTO messages (sender_id, group_id, content) VALUES 
(1, 4, 'Project Alpha kickoff meeting tomorrow at 10 AM.'),
(3, 4, 'I have completed the initial system design.'), 
(5, 4, 'Testing environment is set up and ready.');

-- Social Club messages
INSERT INTO messages (sender_id, group_id, content) VALUES 
(4, 5, 'Friday social event at the rooftop garden!'),
(1, 5, 'Count me in! What should I bring?'),
(2, 5, 'I can bring some snacks and drinks.'),
(3, 5, 'Great idea, I will bring music playlist.'),
(5, 5, 'Looking forward to it!'),
(6, 5, 'Should we make it a potluck style?');

-- ===============================================
-- 6. CONSTRAINT VIOLATION TESTS (COMMENTED OUT)
-- ===============================================

-- This should FAIL (both receiver_id and group_id set)
-- INSERT INTO messages (sender_id, receiver_id, group_id, content) 
-- VALUES (1, 2, 1, 'This will fail!');
-- ERROR: new row for relation "messages" violates check constraint "messages_receiver_or_group_check"

-- This should FAIL (neither receiver_id nor group_id set)
-- INSERT INTO messages (sender_id, content) 
-- VALUES (1, 'This will also fail!');
-- ERROR: new row for relation "messages" violates check constraint "messages_receiver_or_group_check"

-- ===============================================
-- 7. VALIDATION QUERIES
-- ===============================================

-- Query to test message type identification
SELECT 
    id,
    sender_id,
    receiver_id,
    group_id,
    content,
    CASE 
        WHEN group_id IS NOT NULL THEN 'group'
        WHEN receiver_id IS NOT NULL THEN 'direct'
        ELSE 'unknown'
    END as message_type,
    created_at
FROM messages
ORDER BY created_at DESC;

-- ===============================================
-- 8. COMPREHENSIVE TEST QUERIES
-- ===============================================

-- Show all users
SELECT id, username, mobile_no FROM users ORDER BY id;

-- Show all groups with their creators
SELECT 
    g.id,
    g.group_name,
    g.creator_id,
    u.username as creator_name,
    g.created_at
FROM groups g
JOIN users u ON g.creator_id = u.id
ORDER BY g.id;

-- Show group memberships with admin status
SELECT 
    g.group_name,
    u.username,
    gm.is_admin,
    gm.joined_at
FROM group_members gm
JOIN groups g ON gm.group_id = g.id  
JOIN users u ON gm.member_id = u.id
ORDER BY g.id, gm.is_admin DESC, gm.joined_at;

-- Count members and admins per group
SELECT 
    g.group_name,
    COUNT(*) as total_members,
    SUM(CASE WHEN gm.is_admin THEN 1 ELSE 0 END) as admin_count,
    COUNT(*) - SUM(CASE WHEN gm.is_admin THEN 1 ELSE 0 END) as regular_members
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
GROUP BY g.id, g.group_name
ORDER BY total_members DESC;

-- Show direct message conversations
SELECT 
    CONCAT(u1.username, ' → ', u2.username) as conversation,
    m.content,
    m.created_at
FROM messages m
JOIN users u1 ON m.sender_id = u1.id
JOIN users u2 ON m.receiver_id = u2.id  
WHERE m.receiver_id IS NOT NULL
ORDER BY m.created_at;

-- Show group messages with group names
SELECT 
    g.group_name,
    u.username as sender,
    m.content,
    m.created_at
FROM messages m
JOIN groups g ON m.group_id = g.id
JOIN users u ON m.sender_id = u.id
WHERE m.group_id IS NOT NULL
ORDER BY g.group_name, m.created_at;

-- Message activity summary per user
SELECT 
    u.username,
    COUNT(CASE WHEN m.receiver_id IS NOT NULL THEN 1 END) as direct_messages_sent,
    COUNT(CASE WHEN m.group_id IS NOT NULL THEN 1 END) as group_messages_sent,
    COUNT(*) as total_messages_sent
FROM users u
LEFT JOIN messages m ON u.id = m.sender_id
GROUP BY u.id, u.username
ORDER BY total_messages_sent DESC;

-- Messages received per user (direct messages only)
SELECT 
    u.username,
    COUNT(m.id) as direct_messages_received
FROM users u
LEFT JOIN messages m ON u.id = m.receiver_id
GROUP BY u.id, u.username
ORDER BY direct_messages_received DESC;

-- Most active groups by message count
SELECT 
    g.group_name,
    COUNT(m.id) as message_count,
    COUNT(DISTINCT m.sender_id) as unique_senders
FROM groups g
LEFT JOIN messages m ON g.id = m.group_id
GROUP BY g.id, g.group_name
ORDER BY message_count DESC;

-- User participation in groups
SELECT 
    u.username,
    COUNT(gm.group_id) as groups_joined,
    SUM(CASE WHEN gm.is_admin THEN 1 ELSE 0 END) as admin_in_groups
FROM users u
LEFT JOIN group_members gm ON u.id = gm.member_id
GROUP BY u.id, u.username
ORDER BY groups_joined DESC;

-- ===============================================
-- 9. CONSTRAINT VALIDATION TESTS
-- ===============================================

-- Verify no messages violate the constraint
SELECT 
    'Valid messages' as check_type,
    COUNT(*) as count
FROM messages 
WHERE (receiver_id IS NOT NULL AND group_id IS NULL) 
   OR (receiver_id IS NULL AND group_id IS NOT NULL);

-- This should return 0 if constraint is working
SELECT 
    'Invalid messages' as check_type,
    COUNT(*) as count  
FROM messages
WHERE NOT (
    (receiver_id IS NOT NULL AND group_id IS NULL) OR 
    (receiver_id IS NULL AND group_id IS NOT NULL)
);

-- ===============================================
-- 10. PERFORMANCE INDEX TESTING
-- ===============================================

-- Test receiver_id index performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM messages WHERE receiver_id = 1;

-- Test group_id index performance  
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM messages WHERE group_id = 1;

-- Test mixed query performance
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM messages 
WHERE sender_id = 1 
   OR receiver_id = 1 
   OR group_id IN (SELECT group_id FROM group_members WHERE member_id = 1);
ORDER BY created_at DESC;
