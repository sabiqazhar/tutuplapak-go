-- Drop partial unique indexes
DROP INDEX IF EXISTS uq_users_email_not_empty;
DROP INDEX IF EXISTS uq_users_phone_not_empty;

-- Restore original constraints
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_phone_key UNIQUE (phone);

-- Restore duplicates if needed (to mimic old schema)
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_phone_unique UNIQUE (phone);

