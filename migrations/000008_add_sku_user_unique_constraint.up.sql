-- Add unique constraint for SKU per user
-- This ensures that each user can only have one product with the same SKU
ALTER TABLE products ADD CONSTRAINT unique_sku_per_user UNIQUE (user_id, sku);
