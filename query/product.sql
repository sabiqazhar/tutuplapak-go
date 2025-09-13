-- name: CreateProduct :one
INSERT INTO products (user_id, name, category, qty, price, sku, file_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING product_id, user_id, name, category, qty, price, sku, file_id, created_at, updated_at;

-- name: GetProductByID :one
SELECT product_id, user_id, name, category, qty, price, sku, file_id, created_at, updated_at
FROM products
WHERE product_id = $1;

-- name: GetProductBySKUAndUserID :one
SELECT product_id, user_id, name, category, qty, price, sku, file_id, created_at, updated_at
FROM products
WHERE sku = $1 AND user_id = $2;

-- name: GetProductCategoryByName :one
SELECT product_category_id FROM product_category WHERE name = $1;

-- name: ListProducts :many
SELECT
    p.product_id, p.user_id, p.name, pc.name as category_name, p.qty, p.price, p.sku, p.file_id, p.created_at, p.updated_at,
    f.file_uri, f.file_thumnail_uri
FROM products p
         JOIN product_category pc ON p.category = pc.product_category_id
         LEFT JOIN files f on p.file_id = f.id
WHERE
    (sqlc.narg('product_id')::INT IS NULL OR p.product_id = sqlc.narg('product_id')) AND
    (sqlc.narg('sku')::TEXT IS NULL OR p.sku = sqlc.narg('sku')) AND
    (sqlc.narg('category')::TEXT IS NULL OR pc.name = sqlc.narg('category'))
ORDER BY
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'newest' THEN p.created_at END DESC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'oldest' THEN p.created_at END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'cheapest' THEN p.price END ASC,
    CASE WHEN sqlc.narg('sort_by')::TEXT = 'expensive' THEN p.price END DESC,
    p.created_at DESC
    LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');