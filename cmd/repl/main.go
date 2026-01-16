package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Techbite-sudo/pesapal-rdbms/pkg/executor"
	"github.com/Techbite-sudo/pesapal-rdbms/pkg/parser"
	"github.com/Techbite-sudo/pesapal-rdbms/pkg/storage"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

func main() {
	fmt.Println(colorCyan + "╔═══════════════════════════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + "║" + colorReset + "         " + colorPurple + "Pesapal RDBMS - Interactive REPL" + colorReset + "              " + colorCyan + "║" + colorReset)
	fmt.Println(colorCyan + "║" + colorReset + "         " + colorYellow + "Junior Dev Challenge 2026" + colorReset + "                    " + colorCyan + "║" + colorReset)
	fmt.Println(colorCyan + "╚═══════════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()
	fmt.Println(colorBlue + "Type SQL commands and press Enter to execute." + colorReset)
	fmt.Println(colorBlue + "Type 'help' for available commands, 'exit' or 'quit' to exit." + colorReset)
	fmt.Println()

	// Initialize storage
	dataDir := "./data"
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		fmt.Printf(colorRed+"Error initializing storage: %v\n"+colorReset, err)
		os.Exit(1)
	}

	// Initialize executor
	exec := executor.NewExecutor(store)

	// Start REPL
	reader := bufio.NewReader(os.Stdin)
	var multiLineQuery strings.Builder
	inMultiLine := false

	for {
		// Display prompt
		if inMultiLine {
			fmt.Print(colorYellow + "... " + colorReset)
		} else {
			fmt.Print(colorGreen + "pesapal> " + colorReset)
		}

		// Read input
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf(colorRed+"Error reading input: %v\n"+colorReset, err)
			break
		}

		line = strings.TrimSpace(line)

		// Handle empty lines
		if line == "" {
			continue
		}

		// Handle special commands
		if !inMultiLine {
			switch strings.ToLower(line) {
			case "exit", "quit":
				fmt.Println(colorCyan + "Goodbye!" + colorReset)
				return
			case "help":
				printHelp()
				continue
			case "tables":
				listTables(store)
				continue
			case "clear":
				clearScreen()
				continue
			}
		}

		// Build multi-line query
		if inMultiLine {
			multiLineQuery.WriteString(" ")
		}
		multiLineQuery.WriteString(line)

		// Check if query is complete (ends with semicolon)
		if strings.HasSuffix(line, ";") {
			query := multiLineQuery.String()
			multiLineQuery.Reset()
			inMultiLine = false

			// Execute query
			executeQuery(exec, query)
		} else {
			inMultiLine = true
		}
	}
}

func executeQuery(exec *executor.Executor, query string) {
	// Remove trailing semicolon
	query = strings.TrimSuffix(strings.TrimSpace(query), ";")

	// Parse query
	p := parser.NewParser(query)
	stmt, err := p.Parse()
	if err != nil {
		fmt.Printf(colorRed+"Parse error: %v\n"+colorReset, err)
		return
	}

	// Execute statement
	result, err := exec.Execute(stmt)
	if err != nil {
		fmt.Printf(colorRed+"Execution error: %v\n"+colorReset, err)
		return
	}

	// Display result
	if result.Message != "" {
		fmt.Println(colorGreen + result.Message + colorReset)
	} else {
		fmt.Print(result.FormatTable())
	}
	fmt.Println()
}

func printHelp() {
	fmt.Println(colorCyan + "╔═══════════════════════════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + "║" + colorReset + "                    " + colorPurple + "Available Commands" + colorReset + "                    " + colorCyan + "║" + colorReset)
	fmt.Println(colorCyan + "╚═══════════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()
	fmt.Println(colorYellow + "SQL Commands:" + colorReset)
	fmt.Println("  CREATE TABLE <name> (<columns>);")
	fmt.Println("  DROP TABLE <name>;")
	fmt.Println("  INSERT INTO <table> VALUES (<values>);")
	fmt.Println("  SELECT <columns> FROM <table> [WHERE <condition>];")
	fmt.Println("  SELECT <columns> FROM <table1> INNER JOIN <table2> ON <condition>;")
	fmt.Println("  UPDATE <table> SET <column>=<value> [WHERE <condition>];")
	fmt.Println("  DELETE FROM <table> [WHERE <condition>];")
	fmt.Println()
	fmt.Println(colorYellow + "Data Types:" + colorReset)
	fmt.Println("  INTEGER, VARCHAR(size), BOOLEAN, FLOAT")
	fmt.Println()
	fmt.Println(colorYellow + "Constraints:" + colorReset)
	fmt.Println("  PRIMARY KEY, UNIQUE, NOT NULL")
	fmt.Println()
	fmt.Println(colorYellow + "REPL Commands:" + colorReset)
	fmt.Println("  help      - Show this help message")
	fmt.Println("  tables    - List all tables")
	fmt.Println("  clear     - Clear the screen")
	fmt.Println("  exit/quit - Exit the REPL")
	fmt.Println()
	fmt.Println(colorYellow + "Examples:" + colorReset)
	fmt.Println("  CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(100), email VARCHAR(100) UNIQUE);")
	fmt.Println("  INSERT INTO users VALUES (1, 'John Doe', 'john@example.com');")
	fmt.Println("  SELECT * FROM users WHERE id = 1;")
	fmt.Println("  UPDATE users SET name = 'Jane Doe' WHERE id = 1;")
	fmt.Println("  DELETE FROM users WHERE id = 1;")
	fmt.Println()
}

func listTables(store *storage.Storage) {
	tables := store.ListTables()
	if len(tables) == 0 {
		fmt.Println(colorYellow + "No tables found." + colorReset)
		return
	}

	fmt.Println(colorCyan + "╔═══════════════════════════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + "║" + colorReset + "                       " + colorPurple + "Tables" + colorReset + "                           " + colorCyan + "║" + colorReset)
	fmt.Println(colorCyan + "╚═══════════════════════════════════════════════════════════╝" + colorReset)
	fmt.Println()
	for i, table := range tables {
		fmt.Printf("  %d. %s\n", i+1, table)
	}
	fmt.Println()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
