-- Populate product_category table with required categories
INSERT INTO product_category (name) VALUES 
    ('Food'),
    ('Beverage'),
    ('Clothes'),
    ('Furniture'),
    ('Tools')
ON CONFLICT DO NOTHING;
