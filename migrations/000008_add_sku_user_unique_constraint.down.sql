-- Drop unique constraint for SKU per user
ALTER TABLE products DROP CONSTRAINT IF EXISTS unique_sku_per_user;
