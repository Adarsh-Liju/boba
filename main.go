package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

type model struct {
	db        *sql.DB
	host      string
	user      string
	database  string
	state     string // "connect", "menu", "query", "databases", "tables"
	databases []string
	tables    []string
	query     string
	results   [][]string
	columns   []string
	rowCount  int
	err       error
	cursor    int
}

func initialModel() model {
	return model{
		state: "connect",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "connect":
			return m.handleConnect(msg)
		case "menu":
			return m.handleMenu(msg)
		case "query":
			return m.handleQuery(msg)
		case "databases":
			return m.handleDatabases(msg)
		case "tables":
			return m.handleTables(msg)
		}
	case tea.WindowSizeMsg:
		// Handle window resize if needed
	}
	return m, nil
}

func (m model) handleConnect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Try to connect
		dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", m.user, m.query, m.host, m.database)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			m.err = err
			return m, nil
		}

		if err := db.Ping(); err != nil {
			m.err = err
			return m, nil
		}

		m.db = db
		m.state = "menu"
		m.query = ""
		m.err = nil
		return m, nil
	case "backspace":
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
		}
	case "ctrl+c":
		return m, tea.Quit
	default:
		if msg.String() != "" {
			m.query += msg.String()
		}
	}
	return m, nil
}

func (m model) handleMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < 3 {
			m.cursor++
		}
	case "enter":
		switch m.cursor {
		case 0: // Execute Query
			m.state = "query"
			m.query = ""
		case 1: // View Databases
			m.state = "databases"
			m.loadDatabases()
		case 2: // View Tables
			m.state = "tables"
			m.loadTables()
		case 3: // Exit
			return m, tea.Quit
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleQuery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.query != "" {
			m.executeQuery(m.query)
			m.query = ""
		}
	case "escape":
		m.state = "menu"
		m.query = ""
		m.results = nil
		m.columns = nil
		m.rowCount = 0
	case "backspace":
		if len(m.query) > 0 {
			m.query = m.query[:len(m.query)-1]
		}
	case "ctrl+c":
		return m, tea.Quit
	default:
		if msg.String() != "" {
			m.query += msg.String()
		}
	}
	return m, nil
}

func (m model) handleDatabases(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.databases)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.databases) > 0 && m.cursor < len(m.databases) {
			m.database = m.databases[m.cursor]
			// Reconnect with new database
			dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", m.user, m.query, m.host, m.database)
			if db, err := sql.Open("mysql", dsn); err == nil {
				if err := db.Ping(); err == nil {
					m.db.Close()
					m.db = db
				}
			}
			m.state = "menu"
			m.cursor = 0
		}
	case "escape":
		m.state = "menu"
		m.cursor = 0
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleTables(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.tables)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.tables) > 0 && m.cursor < len(m.tables) {
			tableName := m.tables[m.cursor]
			query := fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableName)
			m.executeQuery(query)
			m.state = "query"
		}
	case "escape":
		m.state = "menu"
		m.cursor = 0
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) loadDatabases() {
	rows, err := m.db.Query("SHOW DATABASES")
	if err != nil {
		m.err = err
		return
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err == nil {
			databases = append(databases, db)
		}
	}
	m.databases = databases
	m.cursor = 0
}

func (m *model) loadTables() {
	rows, err := m.db.Query("SHOW TABLES")
	if err != nil {
		m.err = err
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}
	m.tables = tables
	m.cursor = 0
}

func (m *model) executeQuery(query string) {
	rows, err := m.db.Query(query)
	if err != nil {
		m.err = err
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		m.err = err
		return
	}

	var results [][]string
	rowCount := 0
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			m.err = err
			return
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				// Handle different types properly
				switch v := val.(type) {
				case []byte:
					// Convert byte slice to string
					row[i] = strings.ReplaceAll(string(v), "\n", " ")
				case string:
					row[i] = strings.ReplaceAll(v, "\n", " ")
				default:
					// For other types (int, float, etc.), use default formatting
					row[i] = strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " ")
				}
			}
		}
		results = append(results, row)
		rowCount++
	}

	m.columns = columns
	m.results = results
	m.rowCount = rowCount
	m.err = nil
}

func (m model) View() string {
	switch m.state {
	case "connect":
		return m.connectView()
	case "menu":
		return m.menuView()
	case "query":
		return m.queryView()
	case "databases":
		return m.databasesView()
	case "tables":
		return m.tablesView()
	default:
		return "Unknown state"
	}
}

func (m model) connectView() string {
	s := "🔗 MySQL Connection\n\n"
	s += fmt.Sprintf("Host: %s\n", m.host)
	s += fmt.Sprintf("User: %s\n", m.user)
	s += fmt.Sprintf("Database: %s\n", m.database)
	s += fmt.Sprintf("Password: %s\n\n", strings.Repeat("*", len(m.query)))

	if m.err != nil {
		s += fmt.Sprintf("❌ Error: %v\n\n", m.err)
	}

	s += "Press Enter to connect, Ctrl+C to quit\n"
	return s
}

func (m model) menuView() string {
	s := fmt.Sprintf("✅ Connected to MySQL at %s/%s\n\n", m.host, m.database)
	s += "📋 Main Menu\n"
	s += "─────────────\n\n"

	options := []string{"Execute Query", "View Databases", "View Tables", "Exit"}
	for i, option := range options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}

	s += "\nUse ↑↓ to navigate, Enter to select, Ctrl+C to quit\n"
	return s
}

func (m model) queryView() string {
	s := "💬 SQL Query\n"
	s += "───────────\n\n"
	s += fmt.Sprintf("Query: %s\n\n", m.query)

	if m.err != nil {
		s += fmt.Sprintf("❌ Error: %v\n\n", m.err)
	}

	if len(m.results) > 0 {
		s += "📊 Results:\n"
		s += strings.Repeat("─", 50) + "\n"

		// Print header
		for i, col := range m.columns {
			if i > 0 {
				s += " | "
			}
			s += col
		}
		s += "\n"
		s += strings.Repeat("─", 50) + "\n"

		// Print rows (limit to first 10 for display)
		displayRows := m.results
		if len(displayRows) > 10 {
			displayRows = displayRows[:10]
		}

		for _, row := range displayRows {
			for i, cell := range row {
				if i > 0 {
					s += " | "
				}
				if len(cell) > 20 {
					cell = cell[:17] + "..."
				}
				s += cell
			}
			s += "\n"
		}

		if len(m.results) > 10 {
			s += fmt.Sprintf("... and %d more rows\n", len(m.results)-10)
		}

		s += fmt.Sprintf("\n📊 Total: %d rows returned\n", m.rowCount)
	}

	s += "\nPress Enter to execute, Escape to return to menu, Ctrl+C to quit\n"
	return s
}

func (m model) databasesView() string {
	s := "📚 Available Databases\n"
	s += "─────────────────────\n\n"

	if len(m.databases) == 0 {
		s += "No databases found\n"
	} else {
		for i, db := range m.databases {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, db)
		}
	}

	s += "\nUse ↑↓ to navigate, Enter to select, Escape to return\n"
	return s
}

func (m model) tablesView() string {
	s := "📋 Available Tables\n"
	s += "──────────────────\n\n"

	if len(m.tables) == 0 {
		s += "No tables found\n"
	} else {
		for i, table := range m.tables {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, table)
		}
	}

	s += "\nUse ↑↓ to navigate, Enter to view table data, Escape to return\n"
	return s
}

func main() {
	// Get initial connection details
	var host, user, password, database string

	fmt.Print("Host (localhost): ")
	if _, err := fmt.Scanln(&host); err != nil {
		fmt.Printf("Error reading host: %v\n", err)
		os.Exit(1)
	}
	if host == "" {
		host = "localhost"
	}

	fmt.Print("User (root): ")
	if _, err := fmt.Scanln(&user); err != nil {
		fmt.Printf("Error reading user: %v\n", err)
		os.Exit(1)
	}
	if user == "" {
		user = "root"
	}

	fmt.Print("Password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Database: ")
	if _, err := fmt.Scanln(&database); err != nil {
		fmt.Printf("Error reading database: %v\n", err)
		os.Exit(1)
	}

	// Initialize model
	m := initialModel()
	m.host = host
	m.user = user
	m.database = database
	m.query = password // Store password in query field temporarily

	// Create and run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
