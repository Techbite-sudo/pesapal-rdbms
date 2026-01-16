# Pesapal RDBMS - Junior Dev Challenge '26

A simple Relational Database Management System (RDBMS) built from scratch in Go, featuring SQL-like query support, an interactive REPL, and a demonstration web application.

## Project Overview

This project is my submission for the Pesapal Junior Dev Challenge 2026. It demonstrates the implementation of a functional RDBMS with the following capabilities:

- **SQL-like Query Language**: Support for DDL (Data Definition Language) and DML (Data Manipulation Language)
- **CRUD Operations**: Full Create, Read, Update, Delete functionality
- **Data Types**: Support for multiple column data types (INTEGER, VARCHAR, BOOLEAN, FLOAT)
- **Constraints**: PRIMARY KEY and UNIQUE constraints
- **Indexing**: Basic indexing for improved query performance
- **JOIN Operations**: Support for joining multiple tables
- **Interactive REPL**: Command-line interface for direct database interaction
- **Web Application**: Next.js frontend demonstrating practical CRUD usage via REST API

## Architecture

### Backend (Go)
```
pesapal-rdbms/
├── cmd/
│   ├── repl/          # Interactive REPL mode
│   └── server/        # HTTP API server (Fiber)
├── pkg/
│   ├── parser/        # SQL query parser
│   ├── storage/       # File-based storage engine
│   ├── executor/      # Query execution engine
│   └── index/         # Indexing system
```

### Frontend (Next.js)
```
web-app/               # Next.js demonstration application
```

## Features

### Supported SQL Commands

**Data Definition Language (DDL):**
- `CREATE TABLE` - Define new tables with columns and constraints
- `DROP TABLE` - Remove tables from the database

**Data Manipulation Language (DML):**
- `INSERT INTO` - Add new records
- `SELECT` - Query data with filtering and joins
- `UPDATE` - Modify existing records
- `DELETE` - Remove records

**Constraints:**
- `PRIMARY KEY` - Unique identifier for table rows
- `UNIQUE` - Ensure column values are unique

**Joins:**
- `INNER JOIN` - Combine rows from multiple tables

## Getting Started

### Prerequisites
- Go 1.22 or higher
- Node.js 22.x or higher
- pnpm (for Next.js app)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Techbite-sudo/pesapal-rdbms.git
cd pesapal-rdbms
```

2. Install Go dependencies:
```bash
go mod download
```

3. Install web app dependencies:
```bash
cd web-app
pnpm install
```

## Usage

### Interactive REPL Mode

Start the REPL to interact with the database directly:

```bash
go run cmd/repl/main.go
```

Example session:
```sql
> CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100), email VARCHAR(100) UNIQUE);
Table 'users' created successfully.

> INSERT INTO users (id, name, email) VALUES (1, 'John Doe', 'john@example.com');
1 row inserted.

> SELECT * FROM users;
+----+----------+------------------+
| id | name     | email            |
+----+----------+------------------+
| 1  | John Doe | john@example.com |
+----+----------+------------------+
```

### API Server

Start the HTTP API server:

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Web Application

Start the Next.js development server:

```bash
cd web-app
pnpm dev
```

Visit `http://localhost:3000` to interact with the database through the web interface.

## Technical Implementation

### Storage Engine
The database uses a file-based storage system where:
- Each table is stored as a separate file
- Data is persisted in a structured binary format
- Indexes are maintained in separate files for fast lookups

### Query Parser
A custom lexer and parser analyze SQL queries and convert them into an Abstract Syntax Tree (AST) for execution.

### Execution Engine
The executor processes the parsed queries, interacts with the storage layer, and returns results.

### Indexing
B-tree-based indexes are created automatically for PRIMARY KEY and UNIQUE columns to optimize query performance.

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
# Build REPL
go build -o bin/repl cmd/repl/main.go

# Build API server
go build -o bin/server cmd/server/main.go
```

## Acknowledgments

This project was built from scratch as part of the Pesapal Junior Dev Challenge 2026. While AI tools were used to assist in development, all architectural decisions and core implementations are original work.

### Libraries Used
- [Fiber](https://github.com/gofiber/fiber) - Fast HTTP web framework for Go
- [Next.js](https://nextjs.org/) - React framework for the web application
- Standard Go libraries for core functionality

## License

MIT License - See LICENSE file for details

## Author

Built with dedication for the Pesapal Junior Dev Challenge 2026.

---

*"The best way to predict the future is to implement it."*
