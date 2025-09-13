-- name: CreateFile :one
INSERT INTO files (file_uri, file_thumnail_uri, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
    RETURNING id, file_uri, file_thumnail_uri, created_at, updated_at;

-- name: GetFileByID :one
SELECT id, file_uri, file_thumnail_uri, created_at, updated_at
FROM files
WHERE id = $1;

-- name: GetFileByStringID :one
SELECT id, file_uri, file_thumnail_uri, created_at, updated_at
FROM files
WHERE id = $1::integer;
