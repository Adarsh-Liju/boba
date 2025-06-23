package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
)

// ConnectionForm represents the database connection form
type ConnectionForm struct {
	inputs     []textinput.Model
	focus      int
	err        error
	status     string
	connecting bool
	spinner    spinner.Model
}

// ConnectionFormMsg represents messages for the connection form
type ConnectionFormMsg struct {
	db  *sql.DB
	err error
}

type model struct {
	db           *sql.DB
	input        textinput.Model
	table        table.Model
	err          error
	status       string
	showHelp     bool
	showHistory  bool
	queryHistory []string
	currentQuery int
	help         help.Model
	historyList  list.Model
	width        int
	height       int
	// Connection form
	showConnectionForm bool
	connectionForm     ConnectionForm
	// UI state
	loading bool
	spinner spinner.Model
}

type queryResultMsg struct {
	rows [][]string
	cols []string
	err  error
}

type windowSizeMsg struct {
	width  int
	height int
}

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Execute key.Binding
	Quit    key.Binding
	Clear   key.Binding
	Help    key.Binding
	History key.Binding
	Up      key.Binding
	Down    key.Binding
	Select  key.Binding
	Back    key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Execute: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "execute"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		History: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "history"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

var keys = DefaultKeyMap()

// Color palette for better UI
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	successColor   = lipgloss.Color("#10B981") // Green
	textColor      = lipgloss.Color("#6B7280") // Gray
	lightTextColor = lipgloss.Color("#9CA3AF") // Light gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	cardBgColor    = lipgloss.Color("#374151") // Lighter dark gray

	// Enhanced styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	statusStyle = lipgloss.NewStyle().
			Foreground(lightTextColor).
			Italic(true).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successColor)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Background(cardBgColor).
			Padding(0, 1).
			MarginBottom(1)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Bold(true).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lightTextColor).
			Italic(true).
			Background(cardBgColor).
			Padding(1, 2).
			MarginTop(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)
)

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.EnterAltScreen,
		spinner.Tick,
	)
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your SQL query here..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	ti.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	tbl := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Enhanced table styling with modern colors
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor).
		Background(cardBgColor)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(textColor).
		Background(cardBgColor)
	tbl.SetStyles(s)

	// Initialize help with enhanced styling
	h := help.New()
	h.Styles = help.Styles{
		ShortDesc: lipgloss.NewStyle().Foreground(lightTextColor),
		FullDesc:  lipgloss.NewStyle().Foreground(lightTextColor),
		ShortKey:  lipgloss.NewStyle().Foreground(primaryColor).Bold(true),
		FullKey:   lipgloss.NewStyle().Foreground(primaryColor).Bold(true),
		Ellipsis:  lipgloss.NewStyle().Foreground(lightTextColor),
	}

	// Initialize history list with better styling
	historyItems := []list.Item{}
	l := list.New(historyItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Query History"
	l.SetShowHelp(false)
	l.Styles.Title = titleStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)

	// Initialize spinner
	spr := spinner.New()
	spr.Spinner = spinner.Dot
	spr.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return model{
		db:                 nil,
		input:              ti,
		table:              tbl,
		status:             "Welcome to MySQL CLI Client! Please connect to your database.",
		queryHistory:       []string{},
		help:               h,
		historyList:        l,
		showConnectionForm: true,
		connectionForm:     newConnectionForm(),
		spinner:            spr,
	}
}

// Update handles input and query execution
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ConnectionFormMsg:
		if msg.err != nil {
			m.connectionForm.err = msg.err
			m.connectionForm.connecting = false
			m.connectionForm.status = "❌ Connection failed. Please check your details and try again."
		} else {
			m.db = msg.db
			m.showConnectionForm = false
			m.status = "✅ Successfully connected to MySQL database"
			m.input.Focus()
		}
		return m, nil
	case tea.KeyMsg:
		if m.showConnectionForm {
			m.connectionForm, cmd = m.connectionForm.Update(msg)
			return m, cmd
		}

		if m.showHelp {
			switch msg.String() {
			case "esc":
				m.showHelp = false
				m.input.Focus()
			}
			return m, nil
		}

		if m.showHistory {
			switch msg.String() {
			case "esc":
				m.showHistory = false
				m.input.Focus()
			case "enter":
				if m.historyList.SelectedItem() != nil {
					selectedQuery := m.historyList.SelectedItem().(historyItem).query
					m.input.SetValue(selectedQuery)
					m.showHistory = false
					m.input.Focus()
				}
			}
			m.historyList, cmd = m.historyList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			query := strings.TrimSpace(m.input.Value())
			if query != "" {
				// Add to history if not empty
				if len(m.queryHistory) == 0 || m.queryHistory[len(m.queryHistory)-1] != query {
					m.queryHistory = append(m.queryHistory, query)
					// Update history list
					items := make([]list.Item, len(m.queryHistory))
					for i, q := range m.queryHistory {
						items[i] = historyItem{query: q, index: i + 1}
					}
					m.historyList.SetItems(items)
				}
				m.loading = true
				m.status = "⏳ Executing query..."
				cmd := func() tea.Msg {
					rows, cols, err := execQuery(m.db, query)
					return queryResultMsg{rows, cols, err}
				}
				return m, cmd
			}
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.input.SetValue("")
			m.input.Focus()
		case "?":
			m.showHelp = true
			m.input.Blur()
		case "ctrl+h":
			m.showHistory = true
			m.input.Blur()
		case "up":
			if len(m.queryHistory) > 0 {
				if m.currentQuery > 0 {
					m.currentQuery--
				} else {
					m.currentQuery = len(m.queryHistory) - 1
				}
				m.input.SetValue(m.queryHistory[m.currentQuery])
			}
		case "down":
			if len(m.queryHistory) > 0 {
				if m.currentQuery < len(m.queryHistory)-1 {
					m.currentQuery++
				} else {
					m.currentQuery = 0
				}
				m.input.SetValue(m.queryHistory[m.currentQuery])
			}
		}
	case queryResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = "❌ Query execution failed"
		} else {
			m.err = nil
			cols := msg.cols
			m.table = table.New(
				table.WithColumns(makeColumns(cols)),
				table.WithRows(makeRows(msg.rows)),
				table.WithFocused(true),
				table.WithHeight(15),
			)
			// Enhanced table styling with modern colors
			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				BorderBottom(true).
				Bold(true).
				Foreground(primaryColor).
				Background(cardBgColor)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(primaryColor).
				Bold(false)
			s.Cell = s.Cell.
				Foreground(textColor).
				Background(cardBgColor)
			m.table.SetStyles(s)
			m.status = fmt.Sprintf("✅ Query executed successfully! %d rows returned.", len(msg.rows))
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height / 3)
		m.historyList.SetSize(msg.Width-4, msg.Height-10)
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.showHelp && !m.showHistory && !m.showConnectionForm {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.showConnectionForm {
		return m.connectionForm.View()
	}

	if m.showHelp {
		return m.helpView()
	}

	if m.showHistory {
		return m.historyView()
	}

	var s strings.Builder

	// Enhanced header with modern styling
	header := titleStyle.Render("🐬 MySQL CLI Client")
	s.WriteString(header + "\n")

	// Status with icon and better styling
	statusIcon := "🟢"
	if m.err != nil {
		statusIcon = "🔴"
	} else if m.loading {
		statusIcon = "⏳"
	}

	statusText := statusStyle.Render(fmt.Sprintf("%s %s", statusIcon, m.status))
	s.WriteString(statusText + "\n")

	// Error display with enhanced styling
	if m.err != nil {
		errorCard := errorStyle.Render("❌ Error Details: " + m.err.Error())
		s.WriteString(errorCard + "\n")
	}

	// Results section with better styling
	if len(m.table.Rows()) > 0 {
		resultsHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Render("📊 Query Results")
		s.WriteString(resultsHeader + "\n")

		tableCard := cardStyle.Render(m.table.View())
		s.WriteString(tableCard + "\n")
	}

	// Input section with enhanced styling
	inputHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		Background(cardBgColor).
		Padding(1, 2).
		MarginBottom(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Render("💬 SQL Query")
	s.WriteString(inputHeader + "\n")

	inputCard := cardStyle.Render(m.input.View())
	s.WriteString(inputCard + "\n")

	// Help section with enhanced styling
	helpCard := helpStyle.Render("⌨️  Press Enter to execute • Ctrl+C to quit • Esc to clear • ? for help • Ctrl+H for history")
	s.WriteString(helpCard)

	return s.String()
}

func (m model) helpView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL CLI Client - Help & Documentation")
	s.WriteString(header + "\n")

	helpContent := `
🔧 Key Bindings:
  Enter     - Execute SQL query
  Ctrl+C    - Quit application
  Esc       - Clear input field
  ?         - Show/hide this help
  Ctrl+H    - Show query history
  ↑/↓       - Navigate through query history

✨ Features:
  • Interactive SQL query execution
  • Query history with navigation
  • Tabular result display with sorting
  • Error handling and display
  • Responsive layout
  • Database connection management
  • Real-time query execution

🎨 UI Features:
  • Modern terminal interface
  • Color-coded status indicators
  • Smooth animations and transitions
  • Intuitive navigation
  • Professional styling

Press Esc to return to the main interface.
`

	helpCard := cardStyle.Render(helpContent)
	s.WriteString(helpCard)

	return s.String()
}

func (m model) historyView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL CLI Client - Query History")
	s.WriteString(header + "\n")

	historyCard := cardStyle.Render(m.historyList.View())
	s.WriteString(historyCard + "\n")

	helpText := helpStyle.Render("⌨️  Press Enter to select query • Esc to go back • Use ↑↓ to navigate")
	s.WriteString(helpText)

	return s.String()
}

// historyItem represents an item in the query history list
type historyItem struct {
	query string
	index int
}

func (i historyItem) Title() string {
	return fmt.Sprintf("%d. %s", i.index, truncateString(i.query, 60))
}

func (i historyItem) Description() string {
	return ""
}

func (i historyItem) FilterValue() string {
	return i.query
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func execQuery(db *sql.DB, query string) ([][]string, []string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var result [][]string
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, nil, err
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				// Handle different types properly
				switch v := val.(type) {
				case []byte:
					// Convert byte array to string
					row[i] = string(v)
				case string:
					row[i] = v
				case int64:
					row[i] = fmt.Sprintf("%d", v)
				case float64:
					row[i] = fmt.Sprintf("%f", v)
				case bool:
					row[i] = fmt.Sprintf("%t", v)
				default:
					// For any other type, use the default string representation
					row[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		result = append(result, row)
	}

	return result, columns, nil
}

func makeColumns(cols []string) []table.Column {
	columns := make([]table.Column, len(cols))
	for i, col := range cols {
		columns[i] = table.Column{Title: col, Width: 20}
	}
	return columns
}

func makeRows(rows [][]string) []table.Row {
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = row
	}
	return tableRows
}

// newConnectionForm creates a new connection form
func newConnectionForm() ConnectionForm {
	inputs := make([]textinput.Model, 5)

	// Host input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "localhost"
	inputs[0].CharLimit = 50
	inputs[0].Width = 30
	inputs[0].Prompt = "🏠 Host: "
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[0].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[0].PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	// Port input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "3306"
	inputs[1].CharLimit = 5
	inputs[1].Width = 10
	inputs[1].Prompt = "🔌 Port: "
	inputs[1].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[1].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[1].PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	// Username input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "root"
	inputs[2].CharLimit = 50
	inputs[2].Width = 30
	inputs[2].Prompt = "👤 Username: "
	inputs[2].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[2].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[2].PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	// Password input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password"
	inputs[3].CharLimit = 100
	inputs[3].Width = 30
	inputs[3].Prompt = "🔒 Password: "
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[3].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[3].PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	// Database name input
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "database_name"
	inputs[4].CharLimit = 50
	inputs[4].Width = 30
	inputs[4].Prompt = "🗄️  Database: "
	inputs[4].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[4].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	inputs[4].PlaceholderStyle = lipgloss.NewStyle().Foreground(lightTextColor)

	// Focus the first input
	inputs[0].Focus()

	return ConnectionForm{
		inputs:     inputs,
		focus:      0,
		status:     "Enter your MySQL connection details to get started",
		connecting: false,
	}
}

// Update handles the connection form updates
func (cf ConnectionForm) Update(msg tea.Msg) (ConnectionForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle through inputs
			if s == "up" || s == "shift+tab" {
				cf.focus--
			} else {
				cf.focus++
			}

			if cf.focus >= len(cf.inputs) {
				cf.focus = 0
			} else if cf.focus < 0 {
				cf.focus = len(cf.inputs) - 1
			}

			// Update focus
			for i := 0; i <= len(cf.inputs)-1; i++ {
				if i == cf.focus {
					cmd = cf.inputs[i].Focus()
					cf.inputs[i], cmd = cf.inputs[i].Update(cmd)
				} else {
					cf.inputs[i].Blur()
				}
			}
			return cf, cmd
		case "enter":
			// Try to connect
			cf.connecting = true
			cf.status = "⏳ Connecting to database..."
			return cf, cf.connect()
		case "ctrl+c":
			return cf, tea.Quit
		}
	case spinner.TickMsg:
		if cf.connecting {
			cf.spinner, cmd = cf.spinner.Update(msg)
			return cf, cmd
		}
	}

	// Update focused input
	cf.inputs[cf.focus], cmd = cf.inputs[cf.focus].Update(msg)

	return cf, cmd
}

// connect attempts to connect to the database
func (cf ConnectionForm) connect() tea.Cmd {
	return func() tea.Msg {
		// Get values from inputs
		host := cf.inputs[0].Value()
		if host == "" {
			host = "localhost"
		}

		port := cf.inputs[1].Value()
		if port == "" {
			port = "3306"
		}

		user := cf.inputs[2].Value()
		if user == "" {
			user = "root"
		}

		pass := cf.inputs[3].Value()
		dbName := cf.inputs[4].Value()

		// Build connection string
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			user, pass, host, port, dbName)

		// Connect to database
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return ConnectionFormMsg{db: nil, err: err}
		}

		// Test connection
		if err := db.Ping(); err != nil {
			return ConnectionFormMsg{db: nil, err: err}
		}

		return ConnectionFormMsg{db: db, err: nil}
	}
}

// View renders the connection form
func (cf ConnectionForm) View() string {
	var s strings.Builder

	// Header with modern styling
	header := titleStyle.Render("💳 MySQL Database Connection")
	s.WriteString(header + "\n")

	// Status with better styling
	statusCard := cardStyle.Render(
		statusStyle.Render(cf.status),
	)
	s.WriteString(statusCard + "\n")

	// Error display with enhanced styling
	if cf.err != nil {
		errorCard := errorStyle.Render("❌ Connection Error: " + cf.err.Error())
		s.WriteString(errorCard + "\n")
	}

	// Form container with modern styling
	formContainer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Background(cardBgColor).
		Padding(2, 3).
		MarginBottom(2)

	var formContent strings.Builder

	// Form fields with better styling
	for i, input := range cf.inputs {
		fieldContainer := lipgloss.NewStyle().
			MarginBottom(2)

		if i == cf.focus {
			fieldContainer = fieldContainer.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Background(cardBgColor).
				Padding(0, 1)
		}

		formContent.WriteString(fieldContainer.Render(input.View()) + "\n")
	}

	// Connect button with modern styling
	buttonText := "🔗 Connect to Database"
	if cf.connecting {
		buttonText = "⏳ Connecting..."
	}

	buttonCard := cardStyle.Render(
		buttonStyle.Render(buttonText),
	)
	formContent.WriteString(buttonCard + "\n")

	// Help text with better styling
	helpCard := helpStyle.Render("⌨️  Tab/↑↓ to navigate • Enter to connect • Ctrl+C to quit")
	formContent.WriteString(helpCard)

	s.WriteString(formContainer.Render(formContent.String()))

	return s.String()
}

func main() {
	// Start the TUI with better error handling and mouse support
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if err := p.Start(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
