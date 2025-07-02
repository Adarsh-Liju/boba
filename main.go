package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global configuration
var (
	cfgFile  string
	host     string
	port     string
	user     string
	password string
	database string
	verbose  bool
	format   string
)

// Connection holds database connection info
type Connection struct {
	DB      *sql.DB
	Host    string
	Port    string
	User    string
	DB_Name string
}

// Root command
var rootCmd = &cobra.Command{
	Use:   "boba",
	Short: "A modern MySQL CLI client",
	Long: `MySQL CLI is a modern, interactive MySQL client built with Go.
It provides command-line interface for database operations,
query execution, and result visualization.

Examples:
  boba interactive                    # Launch interactive mode
  boba query "SELECT * FROM users"   # Execute single query
  boba status                         # Check connection status
  boba list tables                    # List tables
  boba --help                         # Show help`,
	Version: "1.0.0",
}

// Interactive command - launches interactive mode
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "Launch interactive mode",
	Long:    `Launch the interactive command-line interface for MySQL operations.`,
	Aliases: []string{"i", "shell"},
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		fmt.Printf("✅ Connected to MySQL at %s:%s/%s as %s\n",
			conn.Host, conn.Port, conn.DB_Name, conn.User)
		fmt.Println("Type 'help' for commands, 'exit' or 'quit' to leave")
		fmt.Println(strings.Repeat("─", 50))

		runInteractiveMode(conn)
	},
}

// Query command - execute a single query
var queryCmd = &cobra.Command{
	Use:     "query [SQL]",
	Short:   "Execute a SQL query",
	Long:    `Execute a single SQL query and display the results.`,
	Aliases: []string{"q", "exec"},
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		if verbose {
			fmt.Printf("🔍 Executing query: %s\n", query)
		}

		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		executeQuery(conn.DB, query)
	},
}

// Status command - check database connection
var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Check database connection status",
	Long:    `Check the status of the database connection and display server information.`,
	Aliases: []string{"ping", "test", "info"},
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		showStatus(conn)
	},
}

// List command - list databases or tables
var listCmd = &cobra.Command{
	Use:       "list [databases|tables]",
	Short:     "List databases or tables",
	Long:      `List all databases or tables in the current database.`,
	Aliases:   []string{"ls", "show"},
	ValidArgs: []string{"databases", "tables", "db", "tbl"},
	Args:      cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		target := "tables"
		if len(args) > 0 {
			target = args[0]
		}

		listItems(conn.DB, target)
	},
}

// Describe command - describe table structure
var describeCmd = &cobra.Command{
	Use:     "describe [table_name]",
	Short:   "Describe table structure",
	Long:    `Show the structure of a table including columns, types, and constraints.`,
	Aliases: []string{"desc", "explain"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		describeTable(conn.DB, tableName)
	},
}

// Export command - export query results
var exportCmd = &cobra.Command{
	Use:   "export [query] [filename]",
	Short: "Export query results to file",
	Long:  `Execute a query and export results to CSV file.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		filename := args[1]

		conn, err := setupConnection()
		if err != nil {
			fmt.Printf("❌ Failed to connect: %v\n", err)
			os.Exit(1)
		}
		defer conn.DB.Close()

		exportResults(conn.DB, query, filename)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.boba.yaml)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "MySQL host")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "P", "3306", "MySQL port")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "root", "MySQL username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "MySQL password")
	rootCmd.PersistentFlags().StringVarP(&database, "database", "d", "", "MySQL database name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "output format (table, csv, json)")

	// Bind flags to viper
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	// Add subcommands
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(exportCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".boba")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Println("📄 Using config file:", viper.ConfigFileUsed())
	}

	// Update variables from config
	host = viper.GetString("host")
	port = viper.GetString("port")
	user = viper.GetString("user")
	password = viper.GetString("password")
	database = viper.GetString("database")
	verbose = viper.GetBool("verbose")
	format = viper.GetString("format")
}

// setupConnection creates and tests database connection
func setupConnection() (*Connection, error) {
	// Prompt for missing required fields
	if database == "" {
		fmt.Print("📋 Database name: ")
		fmt.Scanln(&database)
	}

	if password == "" {
		fmt.Print("🔒 Password: ")
		fmt.Scanln(&password)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Connection{
		DB:      db,
		Host:    host,
		Port:    port,
		User:    user,
		DB_Name: database,
	}, nil
}

// runInteractiveMode starts the interactive shell
func runInteractiveMode(conn *Connection) {
	scanner := bufio.NewScanner(os.Stdin)
	var queryBuilder strings.Builder
	var history []string

	for {
		if queryBuilder.Len() == 0 {
			fmt.Print("mysql> ")
		} else {
			fmt.Print("    -> ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		// Handle special commands
		if queryBuilder.Len() == 0 {
			switch strings.ToLower(line) {
			case "exit", "quit", "\\q":
				fmt.Println("👋 Goodbye!")
				return
			case "help", "\\h":
				showHelp()
				continue
			case "status", "\\s":
				showStatus(conn)
				continue
			case "show databases", "\\l":
				listItems(conn.DB, "databases")
				continue
			case "show tables", "\\dt":
				listItems(conn.DB, "tables")
				continue
			case "history", "\\g":
				showHistory(history)
				continue
			case "clear", "\\c":
				queryBuilder.Reset()
				fmt.Println("Query cleared.")
				continue
			}
		}

		// Build multi-line query
		if line != "" {
			if queryBuilder.Len() > 0 {
				queryBuilder.WriteString(" ")
			}
			queryBuilder.WriteString(line)
		}

		// Execute query if it ends with semicolon
		if strings.HasSuffix(line, ";") {
			query := strings.TrimSuffix(queryBuilder.String(), ";")
			query = strings.TrimSpace(query)

			if query != "" {
				history = append(history, query)
				executeQuery(conn.DB, query)
			}

			queryBuilder.Reset()
		}
	}
}

// executeQuery executes a SQL query and displays results
func executeQuery(db *sql.DB, query string) {
	if verbose {
		fmt.Printf("🔍 Executing: %s\n", query)
	}

	// Handle different query types
	queryLower := strings.ToLower(strings.TrimSpace(query))

	if strings.HasPrefix(queryLower, "select") ||
		strings.HasPrefix(queryLower, "show") ||
		strings.HasPrefix(queryLower, "describe") ||
		strings.HasPrefix(queryLower, "desc") {
		executeSelectQuery(db, query)
	} else {
		executeModifyQuery(db, query)
	}
}

// executeSelectQuery handles SELECT queries
func executeSelectQuery(db *sql.DB, query string) {
	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("❌ Failed to get columns: %v\n", err)
		return
	}

	// Collect all rows
	var results [][]string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("❌ Failed to scan row: %v\n", err)
			return
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					row[i] = string(v)
				default:
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		results = append(results, row)
	}

	// Display results based on format
	switch format {
	case "csv":
		displayCSV(columns, results)
	case "json":
		displayJSON(columns, results)
	default:
		displayTable(columns, results)
	}

	fmt.Printf("\n📊 %d rows returned\n", len(results))
}

// executeModifyQuery handles INSERT, UPDATE, DELETE queries
func executeModifyQuery(db *sql.DB, query string) {
	result, err := db.Exec(query)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("⚠️  Query executed but couldn't get affected rows: %v\n", err)
		return
	}

	if lastInsertId, err := result.LastInsertId(); err == nil && lastInsertId > 0 {
		fmt.Printf("✅ Query executed successfully. %d rows affected. Last insert ID: %d\n",
			rowsAffected, lastInsertId)
	} else {
		fmt.Printf("✅ Query executed successfully. %d rows affected.\n", rowsAffected)
	}
}

// displayTable shows results in table format
func displayTable(columns []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("📭 No results returned.")
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(columns))
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add padding
	for i := range colWidths {
		colWidths[i] += 2
		if colWidths[i] > 50 { // Max column width
			colWidths[i] = 50
		}
	}

	// Print top border
	fmt.Print("┌")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			fmt.Print("┬")
		}
	}
	fmt.Println("┐")

	// Print header
	fmt.Print("│")
	for i, col := range columns {
		if len(col) > 48 {
			col = col[:45] + "..."
		}
		fmt.Printf(" %-*s │", colWidths[i]-2, col)
	}
	fmt.Println()

	// Print separator
	fmt.Print("├")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			fmt.Print("┼")
		}
	}
	fmt.Println("┤")

	// Print rows
	for _, row := range rows {
		fmt.Print("│")
		for i, cell := range row {
			if i < len(colWidths) {
				if len(cell) > 48 {
					cell = cell[:45] + "..."
				}
				fmt.Printf(" %-*s │", colWidths[i]-2, cell)
			}
		}
		fmt.Println()
	}

	// Print bottom border
	fmt.Print("└")
	for i, width := range colWidths {
		fmt.Print(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			fmt.Print("┴")
		}
	}
	fmt.Println("┘")
}

// displayCSV shows results in CSV format
func displayCSV(columns []string, rows [][]string) {
	// Print header
	for i, col := range columns {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Print(col)
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Print(",")
			}
			// Escape quotes and wrap in quotes if contains comma
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") {
				cell = "\"" + strings.ReplaceAll(cell, "\"", "\"\"") + "\""
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

// displayJSON shows results in JSON format
func displayJSON(columns []string, rows [][]string) {
	fmt.Println("[")
	for i, row := range rows {
		if i > 0 {
			fmt.Println(",")
		}
		fmt.Print("  {")
		for j, cell := range row {
			if j > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("\"%s\": \"%s\"", columns[j], cell)
		}
		fmt.Print("}")
	}
	if len(rows) > 0 {
		fmt.Println()
	}
	fmt.Println("]")
}

// showStatus displays connection and server information
func showStatus(conn *Connection) {
	fmt.Println("📊 MySQL Connection Status")
	fmt.Println(strings.Repeat("─", 30))
	fmt.Printf("🏠 Host: %s:%s\n", conn.Host, conn.Port)
	fmt.Printf("👤 User: %s\n", conn.User)
	fmt.Printf("🗄️  Database: %s\n", conn.DB_Name)

	// Get server version
	var version string
	if err := conn.DB.QueryRow("SELECT VERSION()").Scan(&version); err == nil {
		fmt.Printf("🔖 Server Version: %s\n", version)
	}

	// Get current time
	var currentTime string
	if err := conn.DB.QueryRow("SELECT NOW()").Scan(&currentTime); err == nil {
		fmt.Printf("⏰ Server Time: %s\n", currentTime)
	}

	// Get connection ID
	var connectionId int
	if err := conn.DB.QueryRow("SELECT CONNECTION_ID()").Scan(&connectionId); err == nil {
		fmt.Printf("🔗 Connection ID: %d\n", connectionId)
	}

	fmt.Println("✅ Connection is active")
}

// listItems lists databases or tables
func listItems(db *sql.DB, target string) {
	var query string
	var title string

	switch strings.ToLower(target) {
	case "databases", "db":
		query = "SHOW DATABASES"
		title = "📚 Available Databases"
	case "tables", "tbl":
		query = "SHOW TABLES"
		title = "📋 Available Tables"
	default:
		fmt.Printf("❌ Invalid target: %s. Use 'databases' or 'tables'\n", target)
		return
	}

	fmt.Println(title)
	fmt.Println(strings.Repeat("─", len(title)))

	executeSelectQuery(db, query)
}

// describeTable shows table structure
func describeTable(db *sql.DB, tableName string) {
	fmt.Printf("📋 Table Structure: %s\n", tableName)
	fmt.Println(strings.Repeat("─", 30))

	query := fmt.Sprintf("DESCRIBE %s", tableName)
	executeSelectQuery(db, query)
}

// exportResults exports query results to CSV file
func exportResults(db *sql.DB, query string, filename string) {
	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("❌ Failed to get columns: %v\n", err)
		return
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	// Write header
	for i, col := range columns {
		if i > 0 {
			file.WriteString(",")
		}
		file.WriteString(col)
	}
	file.WriteString("\n")

	// Write rows
	rowCount := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("❌ Failed to scan row: %v\n", err)
			return
		}

		for i, val := range values {
			if i > 0 {
				file.WriteString(",")
			}

			var cell string
			if val == nil {
				cell = "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					cell = string(v)
				default:
					cell = fmt.Sprintf("%v", v)
				}
			}

			// Escape CSV
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") {
				cell = "\"" + strings.ReplaceAll(cell, "\"", "\"\"") + "\""
			}
			file.WriteString(cell)
		}
		file.WriteString("\n")
		rowCount++
	}

	fmt.Printf("✅ Exported %d rows to %s\n", rowCount, filename)
}

// showHelp displays interactive mode help
func showHelp() {
	fmt.Println("📖 MySQL CLI Interactive Mode Help")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("Commands:")
	fmt.Println("  help, \\h          - Show this help")
	fmt.Println("  status, \\s        - Show connection status")
	fmt.Println("  show databases    - List all databases")
	fmt.Println("  show tables       - List tables in current database")
	fmt.Println("  history, \\g       - Show query history")
	fmt.Println("  clear, \\c         - Clear current query")
	fmt.Println("  exit, quit, \\q    - Exit interactive mode")
	fmt.Println()
	fmt.Println("Query Execution:")
	fmt.Println("  - Type SQL queries and end with semicolon (;)")
	fmt.Println("  - Multi-line queries are supported")
	fmt.Println("  - Use Ctrl+C to interrupt")
	fmt.Println(strings.Repeat("─", 40))
}

// showHistory displays query history
func showHistory(history []string) {
	fmt.Println("📚 Query History")
	fmt.Println(strings.Repeat("─", 20))

	if len(history) == 0 {
		fmt.Println("No queries in history.")
		return
	}

	for i, query := range history {
		fmt.Printf("%3d: %s\n", i+1, query)
	}
	fmt.Println(strings.Repeat("─", 20))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
