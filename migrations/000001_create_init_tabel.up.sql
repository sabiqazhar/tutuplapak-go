-- Table: files (harus dibuat dulu karena direferensikan)
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    file_uri VARCHAR NOT NULL,
    file_thumnail_uri VARCHAR,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: product_category
CREATE TABLE product_category (
    product_category_id SERIAL PRIMARY KEY,
    name VARCHAR
);

-- Table: users — file_id sekarang INTEGER, referensi ke files.id
-- Table: users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    file_id INTEGER REFERENCES files(id) ON DELETE SET NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    bank_account_name VARCHAR(255),
    bank_account_holder VARCHAR(255),
    bank_account_number VARCHAR(50),
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: products — file_id sekarang INTEGER, referensi ke files.id
CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR,
    category INTEGER REFERENCES product_category(product_category_id) ON DELETE SET NULL,
    qty INTEGER,
    price DECIMAL,
    sku VARCHAR,
    file_id INTEGER REFERENCES files(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: purchases
CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    sender_name VARCHAR,
    sender_contact_type VARCHAR,
    sender_contact_detail VARCHAR,
    total INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Table: purchase_item — tambahkan PK jika belum ada
CREATE TABLE purchase_item (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(product_id) ON DELETE CASCADE,
    total DECIMAL,
    qty INTEGER
);

-- Table: payment_detail
CREATE TABLE payment_detail (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    file_id INTEGER REFERENCES files(id) ON DELETE SET NULL 
);

-- Optional: Index untuk optimasi
-- CREATE INDEX idx_users_email ON users(email);
-- CREATE INDEX idx_users_phone ON users(phone);
-- CREATE INDEX idx_users_file_id ON users(file_id);
-- CREATE INDEX idx_products_user_id ON products(user_id);
-- CREATE INDEX idx_products_category ON products(category);
-- CREATE INDEX idx_products_file_id ON products(file_id);
-- CREATE INDEX idx_purchase_item_purchase_id ON purchase_item(purchase_id);
-- CREATE INDEX idx_purchase_item_product_id ON purchase_item(product_id);
-- CREATE INDEX idx_payment_detail_purchase_id ON payment_detail(purchase_id);
-- CREATE INDEX idx_payment_detail_user_id ON payment_detail(user_id);
-- CREATE INDEX idx_payment_detail_file_id ON payment_detail(file_id);
