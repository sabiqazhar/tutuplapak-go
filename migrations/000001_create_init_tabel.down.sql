-- Drop tables in reverse order (child tables first)
DROP TABLE IF EXISTS payment_detail;
DROP TABLE IF EXISTS purchase_item;
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS product_category;
DROP TABLE IF EXISTS files;

-- DROP EXTENSION IF EXISTS citext; -- optional, biasanya tidak perlu di-drop