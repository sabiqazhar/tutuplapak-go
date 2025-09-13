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