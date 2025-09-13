-- Authentication queries
-- name: CreateUser :one
INSERT INTO users (email, password)
VALUES ($1, $2)
    RETURNING id, email, created_at;

-- name: CreateUserWithEmail :one
INSERT INTO users (email, password, phone)
VALUES ($1, $2, $3)
    RETURNING id, email, phone, created_at;

-- name: CreateUserWithPhone :one
INSERT INTO users (phone, password, email)
VALUES ($1, $2, $3)
    RETURNING id, phone, email, created_at;

-- name: GetUserByEmail :one
SELECT id, email, phone, password, created_at
FROM users
WHERE email = $1;

-- name: GetUserByPhone :one
SELECT id, phone, email, password, created_at
FROM users
WHERE phone = $1;

-- Profile management queries
-- name: GetUserByID :one
SELECT id, file_id, email, phone, bank_account_name, bank_account_holder, bank_account_number, password, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET
    file_id = $2,
    bank_account_name = $3,
    bank_account_holder = $4,
    bank_account_number = $5,
    updated_at = NOW()
WHERE id = $1
    RETURNING id, file_id, email, phone, bank_account_name, bank_account_holder, bank_account_number, created_at, updated_at;

-- name: LinkPhoneToUser :one
UPDATE users
SET
    phone = $2,
    updated_at = NOW()
WHERE id = $1
    RETURNING id, file_id, email, phone, bank_account_name, bank_account_holder, bank_account_number, created_at, updated_at;

-- name: LinkEmailToUser :one
UPDATE users
SET
    email = $2,
    updated_at = NOW()
WHERE id = $1
    RETURNING id, file_id, email, phone, bank_account_name, bank_account_holder, bank_account_number, created_at, updated_at;
