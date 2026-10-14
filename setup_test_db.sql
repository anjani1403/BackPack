-- ========================================================
-- BackPack Sample Test Database Setup
-- Database: backpack_shop
-- ========================================================

CREATE DATABASE IF NOT EXISTS backpack_shop;
USE backpack_shop;

-- Drop existing tables to start fresh
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS customers;

-- 1. Customers Table
CREATE TABLE customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    city VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 2. Products Table
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 3. Orders Table
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'completed',
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 4. Order Items Table
CREATE TABLE order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Insert Seed Data
INSERT INTO customers (name, email, city) VALUES
('Rajat Bansal', 'rajat@example.com', 'Delhi'),
('Anjani Kumar', 'anjani@example.com', 'Bangalore'),
('John Doe', 'john@example.com', 'New York'),
('Sarah Connor', 'sarah@example.com', 'Los Angeles');

INSERT INTO products (name, price, stock) VALUES
('MacBook Pro M3', 1999.99, 15),
('Mechanical Keyboard', 89.50, 50),
('Wireless Mouse', 29.99, 120),
('4K Monitor', 399.00, 25);

INSERT INTO orders (customer_id, total_amount, status) VALUES
(1, 2089.49, 'completed'),
(2, 428.99, 'completed'),
(3, 89.50, 'processing');

INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 1, 1, 1999.99),
(1, 2, 1, 89.50),
(2, 4, 1, 399.00),
(2, 3, 1, 29.99),
(3, 2, 1, 89.50);

-- Verify content
SELECT 'Customers count' AS entity, COUNT(*) AS count FROM customers
UNION ALL
SELECT 'Products count', COUNT(*) FROM products
UNION ALL
SELECT 'Orders count', COUNT(*) FROM orders
UNION ALL
SELECT 'Order items count', COUNT(*) FROM order_items;
