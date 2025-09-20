-- Drop existing constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_key;

-- Create partial unique indexes (ignore empty strings)
CREATE UNIQUE INDEX uq_users_email_not_empty
ON users (email)
WHERE email <> '';

CREATE UNIQUE INDEX uq_users_phone_not_empty
ON users (phone)
WHERE phone <> '';

