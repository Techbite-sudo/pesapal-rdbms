# Testing Guide - Pesapal RDBMS

This document provides comprehensive testing instructions for the Pesapal RDBMS project.

## Quick Start Testing

### 1. Test the REPL

Start the interactive REPL:

```bash
go run cmd/repl/main.go
```

Try these commands:

```sql
-- Create a table
CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100), email VARCHAR(100) UNIQUE);

-- Insert data
INSERT INTO users VALUES (1, 'John Doe', 'john@example.com');
INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com');

-- Query data
SELECT * FROM users;
SELECT name, email FROM users WHERE id = 1;

-- Update data
UPDATE users SET name = 'John Updated' WHERE id = 1;

-- Delete data
DELETE FROM users WHERE id = 2;

-- Verify changes
SELECT * FROM users;

-- Create another table for JOIN testing
CREATE TABLE orders (id INTEGER PRIMARY KEY, user_id INTEGER, total FLOAT);
INSERT INTO orders VALUES (1, 1, 99.99);
INSERT INTO orders VALUES (2, 1, 149.50);

-- Test JOIN
SELECT users.name, orders.total FROM users INNER JOIN orders ON users.id = orders.user_id;

-- Clean up
DROP TABLE orders;
DROP TABLE users;
```

### 2. Test the API Server

Start the API server:

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

#### Test with curl:

**Create a table:**
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query":"CREATE TABLE products (id INTEGER PRIMARY KEY, name VARCHAR(200), price FLOAT)"}'
```

**Insert data:**
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query":"INSERT INTO products VALUES (1, \"Laptop\", 999.99)"}'
```

**Query data:**
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query":"SELECT * FROM products"}'
```

**List all tables:**
```bash
curl http://localhost:8080/api/tables
```

**Get table info:**
```bash
curl http://localhost:8080/api/tables/products
```

### 3. Test the Web Application

Start the API server (if not already running):
```bash
go run cmd/server/main.go
```

In a new terminal, start the Next.js development server:
```bash
cd web-app
pnpm dev
```

Visit `http://localhost:3000` in your browser and:

1. Click on example queries to load them
2. Execute CREATE TABLE to create a new table
3. Use INSERT to add data
4. Use SELECT to view data
5. Use UPDATE to modify records
6. Use DELETE to remove records
7. Watch the Tables sidebar update automatically

## Comprehensive Test Suite

### Data Types Testing

```sql
CREATE TABLE test_types (
  id INTEGER PRIMARY KEY,
  name VARCHAR(50),
  active BOOLEAN,
  score FLOAT
);

INSERT INTO test_types VALUES (1, 'Test', true, 95.5);
SELECT * FROM test_types;
```

### Constraints Testing

```sql
-- Test PRIMARY KEY constraint
CREATE TABLE pk_test (id INTEGER PRIMARY KEY);
INSERT INTO pk_test VALUES (1);
INSERT INTO pk_test VALUES (1);  -- Should fail: duplicate primary key

-- Test UNIQUE constraint
CREATE TABLE unique_test (id INTEGER PRIMARY KEY, email VARCHAR(100) UNIQUE);
INSERT INTO unique_test VALUES (1, 'test@example.com');
INSERT INTO unique_test VALUES (2, 'test@example.com');  -- Should fail: duplicate unique value
```

### JOIN Testing

```sql
CREATE TABLE customers (id INTEGER PRIMARY KEY, name VARCHAR(100));
CREATE TABLE orders (id INTEGER PRIMARY KEY, customer_id INTEGER, amount FLOAT);

INSERT INTO customers VALUES (1, 'Alice');
INSERT INTO customers VALUES (2, 'Bob');
INSERT INTO orders VALUES (1, 1, 100.00);
INSERT INTO orders VALUES (2, 1, 200.00);
INSERT INTO orders VALUES (3, 2, 150.00);

-- Test INNER JOIN
SELECT customers.name, orders.amount 
FROM customers 
INNER JOIN orders ON customers.id = orders.customer_id;

-- Test JOIN with WHERE clause
SELECT customers.name, orders.amount 
FROM customers 
INNER JOIN orders ON customers.id = orders.customer_id 
WHERE orders.amount > 100;
```

### WHERE Clause Testing

```sql
CREATE TABLE products (id INTEGER PRIMARY KEY, name VARCHAR(100), price FLOAT, stock INTEGER);

INSERT INTO products VALUES (1, 'Laptop', 999.99, 10);
INSERT INTO products VALUES (2, 'Mouse', 29.99, 50);
INSERT INTO products VALUES (3, 'Keyboard', 79.99, 30);
INSERT INTO products VALUES (4, 'Monitor', 299.99, 15);

-- Test different operators
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE stock < 20;
SELECT * FROM products WHERE price >= 79.99;
SELECT * FROM products WHERE id = 2;
SELECT * FROM products WHERE name != 'Mouse';
```

## Performance Testing

### Indexing Performance

The system automatically creates B-tree indexes for PRIMARY KEY and UNIQUE columns. To test performance:

```sql
-- Create a table with indexed column
CREATE TABLE large_table (id INTEGER PRIMARY KEY, value VARCHAR(100));

-- Insert multiple rows (in REPL, you can run this multiple times)
INSERT INTO large_table VALUES (1, 'value1');
INSERT INTO large_table VALUES (2, 'value2');
-- ... continue inserting

-- Query by primary key (should be fast due to index)
SELECT * FROM large_table WHERE id = 500;
```

## Error Handling Testing

Test various error conditions:

```sql
-- Table doesn't exist
SELECT * FROM nonexistent_table;

-- Invalid syntax
CREATE TABLE;

-- Invalid data type
INSERT INTO users VALUES ('not_an_integer', 'Name', 'email@example.com');

-- Constraint violation
INSERT INTO users VALUES (1, 'Duplicate', 'dup@example.com');  -- Duplicate primary key
```

## Expected Results

All tests should:
- Execute without crashes
- Return appropriate error messages for invalid operations
- Maintain data integrity across operations
- Persist data to disk (check the `data/` directory)
- Handle concurrent requests (for API server)

## Troubleshooting

If tests fail:

1. **Check server logs**: `tail -f server.log`
2. **Verify data directory**: `ls -la data/`
3. **Check port availability**: `lsof -i :8080`
4. **Rebuild binaries**: `go build -o bin/repl cmd/repl/main.go`

## Automated Testing

To run Go tests (if implemented):

```bash
go test ./...
```

## Clean Up

To reset the database:

```bash
rm -rf data/
```

This will delete all tables and data. The storage will be recreated on next use.

---

**Note**: This RDBMS is a demonstration project for the Pesapal Junior Dev Challenge. It implements core database functionality but is not intended for production use.
