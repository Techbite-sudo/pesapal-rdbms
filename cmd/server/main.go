package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/Techbite-sudo/pesapal-rdbms/pkg/executor"
	"github.com/Techbite-sudo/pesapal-rdbms/pkg/parser"
	"github.com/Techbite-sudo/pesapal-rdbms/pkg/storage"
)

var (
	store *storage.Storage
	exec  *executor.Executor
)

// QueryRequest represents a SQL query request
type QueryRequest struct {
	Query string `json:"query"`
}

// QueryResponse represents a SQL query response
type QueryResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message,omitempty"`
	Columns      []string      `json:"columns,omitempty"`
	Rows         [][]interface{} `json:"rows,omitempty"`
	RowsAffected int           `json:"rowsAffected"`
	Error        string        `json:"error,omitempty"`
}

// TableInfo represents table metadata
type TableInfo struct {
	Name    string         `json:"name"`
	Columns []ColumnInfo   `json:"columns"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"dataType"`
	Size       int    `json:"size,omitempty"`
	PrimaryKey bool   `json:"primaryKey"`
	Unique     bool   `json:"unique"`
	NotNull    bool   `json:"notNull"`
}

func main() {
	// Initialize storage
	dataDir := "./data"
	var err error
	store, err = storage.NewStorage(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize executor
	exec = executor.NewExecutor(store)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	app.Get("/", handleRoot)
	app.Get("/api/health", handleHealth)
	app.Post("/api/query", handleQuery)
	app.Get("/api/tables", handleListTables)
	app.Get("/api/tables/:name", handleGetTable)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Pesapal RDBMS API Server starting on port %s", port)
	log.Printf("📊 Data directory: %s", dataDir)
	log.Printf("🔗 API endpoint: http://localhost:%s/api/query", port)
	
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleRoot handles the root endpoint
func handleRoot(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    "Pesapal RDBMS API",
		"version": "1.0.0",
		"challenge": "Junior Dev Challenge 2026",
		"endpoints": fiber.Map{
			"health":      "GET /api/health",
			"query":       "POST /api/query",
			"listTables":  "GET /api/tables",
			"getTable":    "GET /api/tables/:name",
		},
	})
}

// handleHealth handles health check
func handleHealth(c *fiber.Ctx) error {
	tables := store.ListTables()
	return c.JSON(fiber.Map{
		"status": "healthy",
		"tables": len(tables),
	})
}

// handleQuery handles SQL query execution
func handleQuery(c *fiber.Ctx) error {
	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(QueryResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(400).JSON(QueryResponse{
			Success: false,
			Error:   "Query is required",
		})
	}

	// Parse query
	p := parser.NewParser(req.Query)
	stmt, err := p.Parse()
	if err != nil {
		return c.Status(400).JSON(QueryResponse{
			Success: false,
			Error:   fmt.Sprintf("Parse error: %v", err),
		})
	}

	// Execute query
	result, err := exec.Execute(stmt)
	if err != nil {
		return c.Status(500).JSON(QueryResponse{
			Success: false,
			Error:   fmt.Sprintf("Execution error: %v", err),
		})
	}

	// Build response
	response := QueryResponse{
		Success:      true,
		Message:      result.Message,
		Columns:      result.Columns,
		Rows:         result.Rows,
		RowsAffected: result.RowsAffected,
	}

	return c.JSON(response)
}

// handleListTables lists all tables
func handleListTables(c *fiber.Ctx) error {
	tables := store.ListTables()
	
	tableInfos := []TableInfo{}
	for _, tableName := range tables {
		table, err := store.GetTable(tableName)
		if err != nil {
			continue
		}

		columns := []ColumnInfo{}
		for _, col := range table.Schema.Columns {
			columns = append(columns, ColumnInfo{
				Name:       col.Name,
				DataType:   col.DataType.String(),
				Size:       col.Size,
				PrimaryKey: col.PrimaryKey,
				Unique:     col.Unique,
				NotNull:    col.NotNull,
			})
		}

		tableInfos = append(tableInfos, TableInfo{
			Name:    tableName,
			Columns: columns,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"tables":  tableInfos,
	})
}

// handleGetTable gets information about a specific table
func handleGetTable(c *fiber.Ctx) error {
	tableName := c.Params("name")
	
	if !store.TableExists(tableName) {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Table '%s' not found", tableName),
		})
	}

	table, err := store.GetTable(tableName)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	columns := []ColumnInfo{}
	for _, col := range table.Schema.Columns {
		columns = append(columns, ColumnInfo{
			Name:       col.Name,
			DataType:   col.DataType.String(),
			Size:       col.Size,
			PrimaryKey: col.PrimaryKey,
			Unique:     col.Unique,
			NotNull:    col.NotNull,
		})
	}

	// Get row count
	rows := table.SelectRows()

	return c.JSON(fiber.Map{
		"success": true,
		"table": TableInfo{
			Name:    tableName,
			Columns: columns,
		},
		"rowCount": len(rows),
	})
}

// customErrorHandler handles errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
