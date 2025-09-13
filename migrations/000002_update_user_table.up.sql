
-- migration/001_initial.sql

-- Update users table to support both email and phone registration
ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL,
    ADD CONSTRAINT users_email_or_phone_check
        CHECK ((email IS NOT NULL AND email != '') OR (phone IS NOT NULL AND phone != ''));

-- Add unique constraints
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_phone_unique UNIQUE (phone);

-- Make sure at least one of email or phone is provided
-- The constraint above handles this

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);