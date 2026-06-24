-- Migration: 02_seed_data.sql
-- Description: Seed initial team members to the users table if they do not exist
-- Password for all seeded accounts is 'password123' (bcrypt hash: $2a$10$YNzvkNu6by_R5e1oTIeKnAo8KL55whYdjY6h9BPZ3ms)

INSERT INTO users (id, name, email, password_hash, role)
VALUES 
    ('member_rian', 'Rian', 'rian.ux@flowak.com', '$2a$10$YNzvkNu6by_R5e1oTIeKnAo8KL55whYdjY6h9BPZ3ms', 'uiux'),
    ('member_siti', 'Siti', 'siti.fe@flowak.com', '$2a$10$YNzvkNu6by_R5e1oTIeKnAo8KL55whYdjY6h9BPZ3ms', 'frontend'),
    ('member_budi', 'Budi', 'budi.be@flowak.com', '$2a$10$YNzvkNu6by_R5e1oTIeKnAo8KL55whYdjY6h9BPZ3ms', 'backend'),
    ('member_arief', 'Arief', 'arief.be@flowak.com', '$2a$10$YNzvkNu6by_R5e1oTIeKnAo8KL55whYdjY6h9BPZ3ms', 'backend')
ON CONFLICT (email) DO NOTHING;
