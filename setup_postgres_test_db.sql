-- ========================================================
-- BackPack Sample PostgreSQL Test Database Setup
-- Database: backpack_pg_shop
-- ========================================================

-- Connect to default maintenance database to recreate backpack_pg_shop
DROP DATABASE IF EXISTS backpack_pg_shop;
CREATE DATABASE backpack_pg_shop;

\c backpack_pg_shop;

-- Drop tables if they exist
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS customers CASCADE;

-- 1. Customers Table
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    city VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Products Table
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    sku VARCHAR(50) NOT NULL UNIQUE,
    price NUMERIC(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Orders Table
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'completed',
    order_date TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Order Items Table
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL
);

-- Indexes for realistic performance
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_order_items_order ON order_items(order_id);

-- Insert Seed Data
INSERT INTO customers (name, email, city) VALUES
('Rajat Bansal', 'rajat@backpack.io', 'Delhi'),
('Anjani Kumar', 'anjani@backpack.io', 'Bangalore'),
('Alice Smith', 'alice@example.com', 'London'),
('Bob Jones', 'bob@example.com', 'San Francisco');

INSERT INTO products (name, sku, price, stock) VALUES
('MacBook Pro M3 Max', 'LAP-MBP-M3', 3499.99, 10),
('Keychron Q1 Pro Wireless', 'KEY-KQ1-01', 199.50, 45),
('Logitech MX Master 3S', 'MOU-MXM-3S', 99.99, 150),
('Dell UltraSharp 32 4K', 'MON-DEL-32', 799.00, 20);

INSERT INTO orders (customer_id, total_amount, status) VALUES
(1, 3599.98, 'completed'),
(2, 299.49, 'completed'),
(3, 99.99, 'shipped');

INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 1, 1, 3499.99),
(1, 3, 1, 99.99),
(2, 2, 1, 199.50),
(2, 3, 1, 99.99),
(3, 3, 1, 99.99);

-- Verification Query
SELECT 'customers' AS table_name, COUNT(*) AS row_count FROM customers
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'orders', COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items;
