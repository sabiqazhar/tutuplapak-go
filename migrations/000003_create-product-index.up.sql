-- Indexes for the 'products' table
-- This index speeds up filtering by SKU in the WHERE clause.
CREATE INDEX idx_products_sku ON products (sku);

-- This index speeds up the JOIN operation with the 'product_category' table.
CREATE INDEX idx_products_category ON products (category);

-- This index speeds up the LEFT JOIN operation with the 'files' table.
CREATE INDEX idx_products_file_id ON products (file_id);

-- These indexes speed up the sorting operations in the ORDER BY clause.
CREATE INDEX idx_products_created_at ON products (created_at);
CREATE INDEX idx_products_price ON products (price);


-- Index for the 'product_category' table
-- This index speeds up filtering by category name in the WHERE clause.
CREATE INDEX idx_product_category_name ON product_category (name);
