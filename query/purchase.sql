-- name: GetProductForUpdate :one
-- This query now only fetches from the products table, without the JOIN.
SELECT
    product_id, user_id, name, category, qty, price, sku, file_id, created_at, updated_at
FROM products
WHERE product_id = $1;

-- name: GetProductCategoryByID :one
-- New query to fetch a category name by its ID.
SELECT name FROM product_category WHERE product_category_id = $1;

-- name: CreatePurchase :one
INSERT INTO purchases (sender_name, sender_contact_type, sender_contact_detail, total, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
    RETURNING id, created_at;

-- name: CreatePurchaseItem :exec
INSERT INTO purchase_item (purchase_id, product_id, qty, total)
VALUES ($1, $2, $3, $4);

-- name: GetSellerBankDetailsByUserID :one
SELECT bank_account_name, bank_account_holder, bank_account_number
FROM users
WHERE id = $1;

-- name: GetPurchaseByID :one
SELECT id, sender_name, sender_contact_type, sender_contact_detail, total, is_paid, created_at, updated_at
FROM purchases
WHERE id = $1;

-- name: GetPurchaseItemsByPurchaseID :many
SELECT pi.id, pi.purchase_id, pi.product_id, pi.qty, pi.total, p.user_id
FROM purchase_item pi
JOIN products p ON pi.product_id = p.product_id
WHERE pi.purchase_id = $1;

-- name: UpdateProductQuantity :exec
UPDATE products 
SET qty = qty - $2, updated_at = NOW()
WHERE product_id = $1;

-- name: UpdatePurchasePaymentStatus :exec
UPDATE purchases 
SET is_paid = TRUE, updated_at = NOW()
WHERE id = $1;

-- name: CreatePaymentDetail :exec
INSERT INTO payment_detail (purchase_id, user_id, file_id)
VALUES ($1, $2, $3);

