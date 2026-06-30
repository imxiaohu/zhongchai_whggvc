-- Add community terms agreement fields to users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS community_terms_agreed BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS community_terms_agreed_at TIMESTAMP NULL;
