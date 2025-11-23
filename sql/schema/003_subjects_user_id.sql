-- Add user_id to subjects
-- In SQLite, adding a NOT NULL column to an existing table with data is tricky.
-- We will add it as nullable first, or default it.
ALTER TABLE subjects ADD COLUMN user_id TEXT NOT NULL DEFAULT 'legacy_data' REFERENCES users(id) ON DELETE CASCADE;
