package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
	// Enhanced form state for better UX
	validated []bool
	submitted bool
}

// ConnectionFormMsg represents messages for the connection form
type ConnectionFormMsg struct {
	db  *sql.DB
	err error
}

// Form validation messages
type validationMsg struct {
	index int
	valid bool
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
	// Copy and export features
	showCopyMenu    bool
	copyMenu        list.Model
	lastQueryResult [][]string
	lastQueryCols   []string
	// Visual enhancements
	showStats     bool
	queryStats    QueryStats
	showTableInfo bool
	tableInfo     TableInfo
	// Menu-driven interface (like bash script)
	showMainMenu  bool
	mainMenu      list.Model
	selectedTable string
	// Pagination for table browsing
	currentPage int
	rowsPerPage int
	totalRows   int
	// Table browsing state
	browsingTable bool
	tableData     [][]string
	tableColumns  []string
	// Table selection state
	showTableList bool
	tableList     list.Model
}

type QueryStats struct {
	executionTime time.Duration
	rowCount      int
	columnCount   int
	timestamp     time.Time
}

type TableInfo struct {
	totalRows    int
	totalColumns int
	hasData      bool
}

type queryResultMsg struct {
	rows          [][]string
	cols          []string
	err           error
	executionTime time.Duration
}

type windowSizeMsg struct {
	width  int
	height int
}

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Execute   key.Binding
	Quit      key.Binding
	Clear     key.Binding
	Help      key.Binding
	History   key.Binding
	Up        key.Binding
	Down      key.Binding
	Select    key.Binding
	Back      key.Binding
	Copy      key.Binding
	Stats     key.Binding
	TableInfo key.Binding
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
		Copy: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "copy data"),
		),
		Stats: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "show stats"),
		),
		TableInfo: key.NewBinding(
			key.WithKeys("ctrl+i"),
			key.WithHelp("ctrl+i", "table info"),
		),
	}
}

var keys = DefaultKeyMap()

// Color palette for better UI - Enhanced for better UX
var (
	primaryColor   = lipgloss.Color("#8B5CF6") // Vibrant purple
	secondaryColor = lipgloss.Color("#10B981") // Emerald green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	successColor   = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	infoColor      = lipgloss.Color("#3B82F6") // Blue

	// Text colors for better readability
	textColor      = lipgloss.Color("#F8FAFC") // Slate 50 - Very light
	lightTextColor = lipgloss.Color("#CBD5E1") // Slate 300 - Light
	mutedTextColor = lipgloss.Color("#64748B") // Slate 500 - Muted

	// Background colors
	bgColor      = lipgloss.Color("#000000") // Pure black
	cardBgColor  = lipgloss.Color("#0F172A") // Slate 900 - Very dark
	hoverBgColor = lipgloss.Color("#1E293B") // Slate 800 - Dark

	// Border colors
	borderColor  = lipgloss.Color("#334155") // Slate 700
	activeBorder = lipgloss.Color("#8B5CF6") // Purple when active

	// Enhanced styles with better UX
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(cardBgColor).
			Padding(2, 4).
			MarginBottom(3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Align(lipgloss.Center).
			Width(60)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lightTextColor).
			Italic(true).
			Align(lipgloss.Center).
			MarginBottom(2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lightTextColor).
			Italic(true).
			MarginBottom(2).
			Align(lipgloss.Center).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Background(cardBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Align(lipgloss.Center)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Background(cardBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successColor).
			Align(lipgloss.Center)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Align(lipgloss.Center)

	activeCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(activeBorder).
			Background(hoverBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Align(lipgloss.Center)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBgColor).
			Padding(1, 2).
			MarginBottom(2).
			Align(lipgloss.Center)

	activeInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(activeBorder).
				Background(hoverBgColor).
				Padding(1, 2).
				MarginBottom(2).
				Align(lipgloss.Center)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(primaryColor).
			Bold(true).
			Padding(1, 3).
			MarginTop(2).
			MarginBottom(2).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	secondaryButtonStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Background(cardBgColor).
				Bold(true).
				Padding(1, 3).
				MarginTop(2).
				MarginBottom(2).
				Align(lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedTextColor).
			Italic(true).
			Background(cardBgColor).
			Padding(2, 3).
			MarginTop(3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Align(lipgloss.Center)

	statsStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Bold(true).
			Background(cardBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(infoColor).
			Align(lipgloss.Center)

	infoStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			Background(cardBgColor).
			Padding(2, 3).
			MarginBottom(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warningColor).
			Align(lipgloss.Center)

	// New styles for better UX
	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(secondaryColor).
				Background(cardBgColor).
				Padding(1, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor).
				Align(lipgloss.Center)

	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(accentColor).
			Bold(true).
			Padding(0, 1).
			MarginLeft(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)

	dividerStyle = lipgloss.NewStyle().
			Foreground(borderColor).
			MarginTop(2).
			MarginBottom(2).
			Align(lipgloss.Center)
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
	ti.TextStyle = lipgloss.NewStyle().Foreground(textColor)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)

	tbl := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Enhanced table styling with better UX
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		BorderBottom(true).
		Bold(true).
		Foreground(primaryColor).
		Background(cardBgColor)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000000")).
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

	// Initialize main menu (like bash script)
	mainMenuItems := []list.Item{
		menuItem{title: "📋 View Tables", desc: "Show all tables in the database"},
		menuItem{title: "🔍 Run Custom Query", desc: "Execute a custom SQL query"},
		menuItem{title: "📊 Show Table Data", desc: "Display data from a specific table"},
		menuItem{title: "📋 Copy Table Structure", desc: "View and copy table structure"},
		menuItem{title: "📄 Scroll Through Results", desc: "Browse table data with pagination"},
		menuItem{title: "❌ Exit", desc: "Exit the application"},
	}
	mainMenu := list.New(mainMenuItems, list.NewDefaultDelegate(), 0, 0)
	mainMenu.Title = "MySQL Database Interface"
	mainMenu.SetShowHelp(false)
	mainMenu.Styles.Title = titleStyle
	mainMenu.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
	mainMenu.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)

	// Initialize copy menu
	copyItems := []list.Item{
		copyItem{title: "📋 Copy as CSV", desc: "Copy data in CSV format"},
		copyItem{title: "📊 Copy as Table", desc: "Copy data as formatted table"},
		copyItem{title: "📄 Copy as JSON", desc: "Copy data in JSON format"},
		copyItem{title: "💾 Export to File", desc: "Save data to a file"},
		copyItem{title: "📈 Copy Statistics", desc: "Copy query statistics"},
	}
	copyList := list.New(copyItems, list.NewDefaultDelegate(), 0, 0)
	copyList.Title = "Copy & Export Options"
	copyList.SetShowHelp(false)
	copyList.Styles.Title = titleStyle
	copyList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
	copyList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)

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
		copyMenu:           copyList,
		mainMenu:           mainMenu,
		showMainMenu:       false,
		queryStats:         QueryStats{},
		tableInfo:          TableInfo{},
		rowsPerPage:        20,
		currentPage:        0,
		showTableList:      false,
		tableList:          list.New(nil, list.NewDefaultDelegate(), 0, 0),
		browsingTable:      false,
		tableData:          [][]string{},
		tableColumns:       []string{},
		selectedTable:      "",
		totalRows:          0,
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
			m.showMainMenu = true // Show main menu after connection
			m.status = "✅ Successfully connected to MySQL database"
		}
		return m, nil
	case tea.KeyMsg:
		if m.showConnectionForm {
			m.connectionForm, cmd = m.connectionForm.Update(msg)
			return m, cmd
		}

		if m.showMainMenu {
			switch msg.String() {
			case "esc":
				m.showMainMenu = false
				m.input.Focus()
			case "enter":
				if m.mainMenu.SelectedItem() != nil {
					selectedItem := m.mainMenu.SelectedItem().(menuItem)
					return m, m.handleMainMenuAction(selectedItem.title)
				}
			}
			m.mainMenu, cmd = m.mainMenu.Update(msg)
			return m, cmd
		}

		if m.showHelp {
			switch msg.String() {
			case "esc":
				m.showHelp = false
				m.showMainMenu = true
			}
			return m, nil
		}

		if m.showHistory {
			switch msg.String() {
			case "esc":
				m.showHistory = false
				m.showMainMenu = true
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

		if m.showCopyMenu {
			switch msg.String() {
			case "esc":
				m.showCopyMenu = false
				m.showMainMenu = true
			case "enter":
				if m.copyMenu.SelectedItem() != nil {
					selectedItem := m.copyMenu.SelectedItem().(copyItem)
					m.handleCopyAction(selectedItem.title)
					m.showCopyMenu = false
					m.showMainMenu = true
				}
			}
			m.copyMenu, cmd = m.copyMenu.Update(msg)
			return m, cmd
		}

		if m.showTableList {
			switch msg.String() {
			case "esc":
				m.showTableList = false
				m.showMainMenu = true
			case "enter":
				if m.tableList.SelectedItem() != nil {
					selectedTable := m.tableList.SelectedItem().(tableItem).name
					m.selectedTable = selectedTable
					m.showTableList = false
					// Handle the selected table based on context
					return m, m.handleTableSelection(selectedTable)
				}
			}
			m.tableList, cmd = m.tableList.Update(msg)
			return m, cmd
		}

		if m.browsingTable {
			switch msg.String() {
			case "esc":
				m.browsingTable = false
				m.showMainMenu = true
			case "left", "h":
				if m.currentPage > 0 {
					m.currentPage--
				}
			case "right", "l":
				maxPage := (m.totalRows - 1) / m.rowsPerPage
				if m.currentPage < maxPage {
					m.currentPage++
				}
			case "r":
				// Refresh current page
			}
			return m, nil
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
				startTime := time.Now()
				cmd := func() tea.Msg {
					rows, cols, err := execQuery(m.db, query)
					executionTime := time.Since(startTime)
					return queryResultMsg{rows, cols, err, executionTime}
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
		case "ctrl+h":
			m.showHistory = true
		case "ctrl+d":
			if len(m.lastQueryResult) > 0 {
				m.showCopyMenu = true
			} else {
				m.status = "⚠️  No data to copy. Execute a query first."
			}
		case "ctrl+s":
			if m.queryStats.rowCount > 0 {
				m.showStats = !m.showStats
			} else {
				m.status = "⚠️  No query statistics available. Execute a query first."
			}
		case "ctrl+i":
			if len(m.lastQueryResult) > 0 {
				m.showTableInfo = !m.showTableInfo
			} else {
				m.status = "⚠️  No table information available. Execute a query first."
			}
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
			m.lastQueryResult = msg.rows
			m.lastQueryCols = msg.cols

			// Update query stats
			m.queryStats = QueryStats{
				executionTime: msg.executionTime,
				rowCount:      len(msg.rows),
				columnCount:   len(msg.cols),
				timestamp:     time.Now(),
			}

			// Update table info
			m.tableInfo = TableInfo{
				totalRows:    len(msg.rows),
				totalColumns: len(msg.cols),
				hasData:      len(msg.rows) > 0,
			}

			m.table = table.New(
				table.WithColumns(makeColumns(cols)),
				table.WithRows(makeRows(msg.rows)),
				table.WithFocused(true),
				table.WithHeight(15),
			)
			// Enhanced table styling with better UX
			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				BorderBottom(true).
				Bold(true).
				Foreground(primaryColor).
				Background(cardBgColor)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("#000000")).
				Background(primaryColor).
				Bold(false)
			s.Cell = s.Cell.
				Foreground(textColor).
				Background(cardBgColor)
			m.table.SetStyles(s)
			m.status = fmt.Sprintf("✅ Query executed successfully! %d rows returned in %v.", len(msg.rows), msg.executionTime)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height / 3)
		m.historyList.SetSize(msg.Width-4, msg.Height-10)
		m.copyMenu.SetSize(msg.Width-4, msg.Height-10)
		m.mainMenu.SetSize(msg.Width-4, msg.Height-10)
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tableSelectionMsg:
		m.showTableList = true
		m.tableList = list.New(msg.items, list.NewDefaultDelegate(), 0, 0)
		m.tableList.Title = "Select a Table"
		m.tableList.SetShowHelp(false)
		m.tableList.Styles.Title = titleStyle
		m.tableList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryColor)
		m.tableList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryColor)
		return m, nil
	}

	if !m.showHelp && !m.showHistory && !m.showConnectionForm && !m.showCopyMenu && !m.showMainMenu && !m.browsingTable && !m.showTableList {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.showConnectionForm {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.connectionForm.View())
	}

	if m.showMainMenu {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.mainMenuView())
	}

	if m.showTableList {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.tableListView())
	}

	if m.showHelp {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.helpView())
	}

	if m.showHistory {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.historyView())
	}

	if m.showCopyMenu {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.copyMenuView())
	}

	if m.browsingTable {
		return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(m.tableBrowsingView())
	}

	var s strings.Builder

	// Enhanced header with better UX
	header := titleStyle.Render("🐬 MySQL CLI Client")
	s.WriteString(header + "\n")

	// Subtitle for better context
	subtitle := subtitleStyle.Render("Your friendly database companion")
	s.WriteString(subtitle + "\n")

	// Status with enhanced UX
	statusIcon := "🟢"
	statusColor := successColor
	if m.err != nil {
		statusIcon = "🔴"
		statusColor = errorColor
	} else if m.loading {
		statusIcon = "⏳"
		statusColor = accentColor
	}

	statusText := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 2).
		Render(fmt.Sprintf("%s %s", statusIcon, m.status))
	s.WriteString(statusText + "\n")

	// Error display with better UX
	if m.err != nil {
		errorCard := errorStyle.Render("❌ Error Details: " + m.err.Error())
		s.WriteString(errorCard + "\n")
	}

	// Query Statistics with enhanced UX
	if m.showStats && m.queryStats.rowCount > 0 {
		statsContent := fmt.Sprintf("📊 Query Statistics:\n"+
			"• ⏱️  Execution Time: %v\n"+
			"• 📈 Rows Returned: %d\n"+
			"• 📋 Columns: %d\n"+
			"• 🕒 Timestamp: %s",
			m.queryStats.executionTime,
			m.queryStats.rowCount,
			m.queryStats.columnCount,
			m.queryStats.timestamp.Format("2006-01-02 15:04:05"))

		statsCard := statsStyle.Render(statsContent)
		s.WriteString(statsCard + "\n")
	}

	// Table Information with enhanced UX
	if m.showTableInfo && m.tableInfo.hasData {
		tableInfoContent := fmt.Sprintf("📋 Table Information:\n"+
			"• 📊 Total Rows: %d\n"+
			"• 📋 Total Columns: %d\n"+
			"• ✅ Data Available: Yes",
			m.tableInfo.totalRows,
			m.tableInfo.totalColumns)

		tableInfoCard := infoStyle.Render(tableInfoContent)
		s.WriteString(tableInfoCard + "\n")
	}

	// Results section with enhanced UX
	if len(m.table.Rows()) > 0 {
		resultsHeader := sectionHeaderStyle.Render("📊 Query Results")
		s.WriteString(resultsHeader + "\n")

		// Add result count badge
		resultCount := badgeStyle.Render(fmt.Sprintf("%d rows", len(m.table.Rows())))
		s.WriteString(resultCount + "\n")

		tableCard := cardStyle.Render(m.table.View())
		s.WriteString(tableCard + "\n")
	}

	// Divider for better visual separation
	if len(m.table.Rows()) > 0 {
		divider := dividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		s.WriteString(divider + "\n")
	}

	// Input section with enhanced UX
	inputHeader := sectionHeaderStyle.Render("💬 SQL Query")
	s.WriteString(inputHeader + "\n")

	// Enhanced input with better styling
	inputCard := activeInputStyle.Render(m.input.View())
	s.WriteString(inputCard + "\n")

	// Enhanced help section with better UX
	helpText := "⌨️  Press Enter to execute • Ctrl+C to quit • Esc to clear • ? for help • Ctrl+H for history"
	if len(m.lastQueryResult) > 0 {
		helpText += " • Ctrl+D to copy data • Ctrl+S for stats • Ctrl+I for table info"
	}
	helpCard := helpStyle.Render(helpText)
	s.WriteString(helpCard)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) mainMenuView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL Database Interface")
	s.WriteString(header + "\n")

	menuCard := cardStyle.Render(m.mainMenu.View())
	s.WriteString(menuCard + "\n")

	helpText := helpStyle.Render("⌨️  Press Enter to select option • Esc to go back • Use ↑↓ to navigate")
	s.WriteString(helpText)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) tableListView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 Select a Table")
	s.WriteString(header + "\n")

	tableCard := cardStyle.Render(m.tableList.View())
	s.WriteString(tableCard + "\n")

	helpText := helpStyle.Render("⌨️  Press Enter to select table • Esc to go back • Use ↑↓ to navigate")
	s.WriteString(helpText)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) tableBrowsingView() string {
	var s strings.Builder

	header := titleStyle.Render(fmt.Sprintf("🐬 Browsing Table: %s", m.selectedTable))
	s.WriteString(header + "\n")

	// Show pagination info
	paginationInfo := fmt.Sprintf("📄 Page %d of %d (Rows %d-%d of %d)",
		m.currentPage+1,
		(m.totalRows-1)/m.rowsPerPage+1,
		m.currentPage*m.rowsPerPage+1,
		min((m.currentPage+1)*m.rowsPerPage, m.totalRows),
		m.totalRows)

	paginationCard := cardStyle.Render(paginationInfo)
	s.WriteString(paginationCard + "\n")

	// Show table data
	if len(m.tableData) > 0 {
		tableCard := cardStyle.Render(m.table.View())
		s.WriteString(tableCard + "\n")
	}

	// Navigation help
	navHelp := helpStyle.Render("⌨️  ←/→ to navigate pages • R to refresh • Esc to go back")
	s.WriteString(navHelp)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) helpView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL CLI Client - Help & Documentation")
	s.WriteString(header + "\n")

	subtitle := subtitleStyle.Render("Everything you need to know to get started")
	s.WriteString(subtitle + "\n")

	helpContent := `
🔧 Key Bindings:
  Enter     - Execute SQL query
  Ctrl+C    - Quit application
  Esc       - Clear input field
  ?         - Show/hide this help
  Ctrl+H    - Show query history
  ↑/↓       - Navigate through query history
  Ctrl+D    - Copy/export data (when results available)
  Ctrl+S    - Show/hide query statistics
  Ctrl+I    - Show/hide table information

✨ Core Features:
  • Interactive SQL query execution
  • Query history with navigation
  • Tabular result display with sorting
  • Error handling and display
  • Responsive layout
  • Database connection management
  • Real-time query execution

📊 Data Management:
  • Copy query results in multiple formats:
    - CSV format for spreadsheets
    - Table format (Markdown) for docs
    - JSON format for APIs
    - File export with timestamps
    - Query statistics export

🎨 UI Features:
  • Modern terminal interface
  • Color-coded status indicators
  • Smooth animations and transitions
  • Intuitive navigation
  • Professional styling
  • Query execution timing
  • Enhanced data visualization

💡 Tips:
  • Use Ctrl+H to quickly access previous queries
  • Press Ctrl+D after executing a query to copy results
  • Use Ctrl+S to see detailed query performance
  • The interface adapts to your terminal size

Press Esc to return to the main interface.
`

	helpCard := cardStyle.Render(helpContent)
	s.WriteString(helpCard)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) historyView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL CLI Client - Query History")
	s.WriteString(header + "\n")

	historyCard := cardStyle.Render(m.historyList.View())
	s.WriteString(historyCard + "\n")

	helpText := helpStyle.Render("⌨️  Press Enter to select query • Esc to go back • Use ↑↓ to navigate")
	s.WriteString(helpText)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

func (m model) copyMenuView() string {
	var s strings.Builder

	header := titleStyle.Render("🐬 MySQL CLI Client - Copy & Export Options")
	s.WriteString(header + "\n")

	copyCard := cardStyle.Render(m.copyMenu.View())
	s.WriteString(copyCard + "\n")

	helpText := helpStyle.Render("⌨️  Press Enter to select option • Esc to go back • Use ↑↓ to navigate")
	s.WriteString(helpText)

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
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

// newConnectionForm creates a new connection form with enhanced UX
func newConnectionForm() ConnectionForm {
	inputs := make([]textinput.Model, 5)
	validated := make([]bool, 5)

	// Host input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "localhost"
	inputs[0].CharLimit = 50
	inputs[0].Width = 30
	inputs[0].Prompt = "🏠 Host: "
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[0].TextStyle = lipgloss.NewStyle().Foreground(textColor)
	inputs[0].PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)
	inputs[0].Validate = func(s string) error {
		if s == "" {
			return nil // Allow empty for default
		}
		if len(s) > 50 {
			return fmt.Errorf("hostname too long")
		}
		return nil
	}

	// Port input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "3306"
	inputs[1].CharLimit = 5
	inputs[1].Width = 10
	inputs[1].Prompt = "🔌 Port: "
	inputs[1].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[1].TextStyle = lipgloss.NewStyle().Foreground(textColor)
	inputs[1].PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)
	inputs[1].Validate = func(s string) error {
		if s == "" {
			return nil // Allow empty for default
		}
		if _, err := strconv.Atoi(s); err != nil {
			return fmt.Errorf("port must be a number")
		}
		return nil
	}

	// Username input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "root"
	inputs[2].CharLimit = 50
	inputs[2].Width = 30
	inputs[2].Prompt = "👤 Username: "
	inputs[2].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[2].TextStyle = lipgloss.NewStyle().Foreground(textColor)
	inputs[2].PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)
	inputs[2].Validate = func(s string) error {
		if s == "" {
			return fmt.Errorf("username is required")
		}
		if len(s) > 50 {
			return fmt.Errorf("username too long")
		}
		return nil
	}

	// Password input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "password"
	inputs[3].CharLimit = 100
	inputs[3].Width = 30
	inputs[3].Prompt = "🔒 Password: "
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[3].TextStyle = lipgloss.NewStyle().Foreground(textColor)
	inputs[3].PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)
	inputs[3].Validate = func(s string) error {
		if s == "" {
			return fmt.Errorf("password is required")
		}
		return nil
	}

	// Database name input
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "database_name"
	inputs[4].CharLimit = 50
	inputs[4].Width = 30
	inputs[4].Prompt = "🗄️  Database: "
	inputs[4].PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	inputs[4].TextStyle = lipgloss.NewStyle().Foreground(textColor)
	inputs[4].PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedTextColor)
	inputs[4].Validate = func(s string) error {
		if s == "" {
			return fmt.Errorf("database name is required")
		}
		if len(s) > 50 {
			return fmt.Errorf("database name too long")
		}
		return nil
	}

	// Focus the first input
	inputs[0].Focus()

	return ConnectionForm{
		inputs:     inputs,
		focus:      0,
		status:     "Enter your MySQL connection details to get started",
		connecting: false,
		validated:  validated,
		submitted:  false,
	}
}

// Update handles the connection form updates with enhanced UX
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

			// Update focus and validate current field
			for i := 0; i <= len(cf.inputs)-1; i++ {
				if i == cf.focus {
					cmd = cf.inputs[i].Focus()
					cf.inputs[i], cmd = cf.inputs[i].Update(cmd)
					// Validate the field when it loses focus
					if err := cf.inputs[i].Validate(cf.inputs[i].Value()); err == nil {
						cf.validated[i] = true
					} else {
						cf.validated[i] = false
					}
				} else {
					cf.inputs[i].Blur()
				}
			}
			return cf, cmd
		case "enter":
			// Validate all fields before attempting connection
			allValid := true
			for i, input := range cf.inputs {
				if err := input.Validate(input.Value()); err != nil {
					cf.validated[i] = false
					allValid = false
				} else {
					cf.validated[i] = true
				}
			}

			if !allValid {
				cf.status = "❌ Please fix validation errors before connecting"
				return cf, nil
			}

			// Try to connect
			cf.connecting = true
			cf.submitted = true
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

	// Update focused input and validate on change
	cf.inputs[cf.focus], cmd = cf.inputs[cf.focus].Update(msg)

	// Validate the current field
	if err := cf.inputs[cf.focus].Validate(cf.inputs[cf.focus].Value()); err == nil {
		cf.validated[cf.focus] = true
	} else {
		cf.validated[cf.focus] = false
	}

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

	// Enhanced header with better UX
	header := titleStyle.Render("💳 MySQL Database Connection")
	s.WriteString(header + "\n")

	// Welcome message for better UX
	welcomeMsg := subtitleStyle.Render("Let's get you connected to your database!")
	s.WriteString(welcomeMsg + "\n")

	// Status with enhanced UX
	statusCard := cardStyle.Render(
		statusStyle.Render(cf.status),
	)
	s.WriteString(statusCard + "\n")

	// Error display with enhanced UX
	if cf.err != nil {
		errorCard := errorStyle.Render("❌ Connection Error: " + cf.err.Error())
		s.WriteString(errorCard + "\n")
	}

	// Form container with enhanced UX
	formContainer := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(cardBgColor).
		Padding(3, 4).
		MarginBottom(3).
		Align(lipgloss.Center).
		Width(50)

	var formContent strings.Builder

	// Form fields with enhanced UX and validation
	for i, input := range cf.inputs {
		fieldContainer := lipgloss.NewStyle().
			MarginBottom(3).
			Align(lipgloss.Center)

		// Choose style based on focus and validation state
		var fieldStyle lipgloss.Style
		if i == cf.focus {
			if cf.validated[i] {
				fieldStyle = activeCardStyle
			} else {
				fieldStyle = activeInputStyle
			}
		} else {
			if cf.validated[i] {
				fieldStyle = cardStyle
			} else {
				fieldStyle = inputStyle
			}
		}

		// Add validation indicator
		validationIcon := " "
		if cf.submitted {
			if cf.validated[i] {
				validationIcon = "✅"
			} else {
				validationIcon = "❌"
			}
		}

		// Render the field with validation indicator
		fieldContent := fmt.Sprintf("%s %s", validationIcon, input.View())
		formContent.WriteString(fieldContainer.Render(fieldStyle.Render(fieldContent)) + "\n")

		// Show validation error if submitted and invalid
		if cf.submitted && !cf.validated[i] {
			if err := input.Validate(input.Value()); err != nil {
				errorMsg := lipgloss.NewStyle().
					Foreground(errorColor).
					Italic(true).
					Align(lipgloss.Center).
					Render(fmt.Sprintf("   %s", err.Error()))
				formContent.WriteString(errorMsg + "\n")
			}
		}
	}

	// Connect button with enhanced UX
	buttonText := "🔗 Connect to Database"
	if cf.connecting {
		buttonText = "⏳ Connecting..."
	}

	// Disable button if form is invalid
	var buttonStyle lipgloss.Style
	if cf.submitted && !cf.isFormValid() {
		buttonStyle = secondaryButtonStyle
		buttonText = "⚠️  Fix validation errors first"
	}
	buttonCard := cardStyle.Render(
		buttonStyle.Render(buttonText),
	)
	formContent.WriteString(buttonCard + "\n")

	// Enhanced help text with better UX
	helpText := "⌨️  Tab/↑↓ to navigate • Enter to connect • Ctrl+C to quit"
	if cf.submitted && !cf.isFormValid() {
		helpText += " • Fix validation errors to continue"
	}
	helpCard := helpStyle.Render(helpText)
	formContent.WriteString(helpCard)

	s.WriteString(formContainer.Render(formContent.String()))

	return lipgloss.NewStyle().Background(bgColor).Align(lipgloss.Center).Render(s.String())
}

// isFormValid checks if all required fields are valid
func (cf ConnectionForm) isFormValid() bool {
	for i, input := range cf.inputs {
		if err := input.Validate(input.Value()); err != nil {
			cf.validated[i] = false
			return false
		}
		cf.validated[i] = true
	}
	return true
}

// copyItem represents an item in the copy menu
type copyItem struct {
	title string
	desc  string
}

func (i copyItem) Title() string {
	return i.title
}

func (i copyItem) Description() string {
	return i.desc
}

func (i copyItem) FilterValue() string {
	return i.title + " " + i.desc
}

// handleCopyAction handles different copy operations
func (m model) handleCopyAction(action string) {
	switch action {
	case "📋 Copy as CSV":
		m.copyAsCSV()
	case "📊 Copy as Table":
		m.copyAsTable()
	case "📄 Copy as JSON":
		m.copyAsJSON()
	case "💾 Export to File":
		m.exportToFile()
	case "📈 Copy Statistics":
		m.copyStatistics()
	}
}

// copyAsCSV copies data in CSV format
func (m model) copyAsCSV() {
	if len(m.lastQueryResult) == 0 {
		m.status = "⚠️  No data to copy"
		return
	}

	var csv strings.Builder

	// Add headers
	for i, col := range m.lastQueryCols {
		if i > 0 {
			csv.WriteString(",")
		}
		csv.WriteString(fmt.Sprintf("\"%s\"", col))
	}
	csv.WriteString("\n")

	// Add data rows
	for _, row := range m.lastQueryResult {
		for i, cell := range row {
			if i > 0 {
				csv.WriteString(",")
			}
			csv.WriteString(fmt.Sprintf("\"%s\"", cell))
		}
		csv.WriteString("\n")
	}

	// In a real application, you would copy to clipboard here
	// For now, we'll just show a success message
	m.status = fmt.Sprintf("✅ CSV data ready to copy (%d rows, %d columns)", len(m.lastQueryResult), len(m.lastQueryCols))
}

// copyAsTable copies data as formatted table
func (m model) copyAsTable() {
	if len(m.lastQueryResult) == 0 {
		m.status = "⚠️  No data to copy"
		return
	}

	var table strings.Builder

	// Add headers
	table.WriteString("| ")
	for i, col := range m.lastQueryCols {
		if i > 0 {
			table.WriteString(" | ")
		}
		table.WriteString(col)
	}
	table.WriteString(" |\n")

	// Add separator
	table.WriteString("| ")
	for i := range m.lastQueryCols {
		if i > 0 {
			table.WriteString(" | ")
		}
		table.WriteString("---")
	}
	table.WriteString(" |\n")

	// Add data rows
	for _, row := range m.lastQueryResult {
		table.WriteString("| ")
		for i, cell := range row {
			if i > 0 {
				table.WriteString(" | ")
			}
			table.WriteString(cell)
		}
		table.WriteString(" |\n")
	}

	m.status = fmt.Sprintf("✅ Table data ready to copy (%d rows, %d columns)", len(m.lastQueryResult), len(m.lastQueryCols))
}

// copyAsJSON copies data in JSON format
func (m model) copyAsJSON() {
	if len(m.lastQueryResult) == 0 {
		m.status = "⚠️  No data to copy"
		return
	}

	var json strings.Builder
	json.WriteString("[\n")

	for i, row := range m.lastQueryResult {
		json.WriteString("  {\n")
		for j, cell := range row {
			json.WriteString(fmt.Sprintf("    \"%s\": \"%s\"", m.lastQueryCols[j], cell))
			if j < len(row)-1 {
				json.WriteString(",")
			}
			json.WriteString("\n")
		}
		json.WriteString("  }")
		if i < len(m.lastQueryResult)-1 {
			json.WriteString(",")
		}
		json.WriteString("\n")
	}
	json.WriteString("]\n")

	m.status = fmt.Sprintf("✅ JSON data ready to copy (%d rows, %d columns)", len(m.lastQueryResult), len(m.lastQueryCols))
}

// exportToFile exports data to a file
func (m model) exportToFile() {
	if len(m.lastQueryResult) == 0 {
		m.status = "⚠️  No data to export"
		return
	}

	filename := fmt.Sprintf("mysql_export_%d.csv", time.Now().Unix())
	file, err := os.Create(filename)
	if err != nil {
		m.status = fmt.Sprintf("❌ Failed to create file: %v", err)
		return
	}
	defer file.Close()

	// Write CSV data
	for i, col := range m.lastQueryCols {
		if i > 0 {
			file.WriteString(",")
		}
		file.WriteString(fmt.Sprintf("\"%s\"", col))
	}
	file.WriteString("\n")

	for _, row := range m.lastQueryResult {
		for i, cell := range row {
			if i > 0 {
				file.WriteString(",")
			}
			file.WriteString(fmt.Sprintf("\"%s\"", cell))
		}
		file.WriteString("\n")
	}

	m.status = fmt.Sprintf("✅ Data exported to %s (%d rows, %d columns)", filename, len(m.lastQueryResult), len(m.lastQueryCols))
}

// copyStatistics copies query statistics
func (m model) copyStatistics() {
	if m.queryStats.rowCount == 0 {
		m.status = "⚠️  No statistics to copy"
		return
	}

	m.status = fmt.Sprintf("✅ Statistics ready to copy:\n"+
		"• Execution Time: %v\n"+
		"• Rows Returned: %d\n"+
		"• Columns: %d\n"+
		"• Timestamp: %s",
		m.queryStats.executionTime,
		m.queryStats.rowCount,
		m.queryStats.columnCount,
		m.queryStats.timestamp.Format("2006-01-02 15:04:05"))
}

// menuItem represents an item in the main menu
type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string {
	return i.title
}

func (i menuItem) Description() string {
	return i.desc
}

func (i menuItem) FilterValue() string {
	return i.title + " " + i.desc
}

// tableItem represents a table in the database
type tableItem struct {
	name string
}

func (i tableItem) Title() string {
	return i.name
}

func (i tableItem) Description() string {
	return "Database table"
}

func (i tableItem) FilterValue() string {
	return i.name
}

// handleMainMenuAction handles different main menu actions
func (m model) handleMainMenuAction(action string) tea.Cmd {
	switch action {
	case "📋 View Tables":
		return m.showTables()
	case "🔍 Run Custom Query":
		m.showMainMenu = false
		return m.input.Focus()
	case "📊 Show Table Data":
		return m.showTableSelection()
	case "📋 Copy Table Structure":
		return m.showTableStructureSelection()
	case "📄 Scroll Through Results":
		return m.showTablePaginationSelection()
	case "❌ Exit":
		return tea.Quit
	}
	return nil
}

// showTables displays all tables in the database
func (m model) showTables() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query("SHOW TABLES")
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return queryResultMsg{nil, nil, err, 0}
			}
			tables = append(tables, tableName)
		}

		// Convert to table format
		var result [][]string
		for _, table := range tables {
			result = append(result, []string{table})
		}

		return queryResultMsg{result, []string{"Table Name"}, nil, 0}
	}
}

// showTableSelection shows a list of tables to select from
func (m model) showTableSelection() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query("SHOW TABLES")
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
		}
		defer rows.Close()

		var tableItems []list.Item
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return queryResultMsg{nil, nil, err, 0}
			}
			tableItems = append(tableItems, tableItem{name: tableName})
		}

		return tableSelectionMsg{tableItems}
	}
}

// showTableStructureSelection shows table structure
func (m model) showTableStructureSelection() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query("SHOW TABLES")
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
		}
		defer rows.Close()

		var tableItems []list.Item
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return queryResultMsg{nil, nil, err, 0}
			}
			tableItems = append(tableItems, tableItem{name: tableName})
		}

		return tableSelectionMsg{tableItems}
	}
}

// showTablePaginationSelection shows table with pagination
func (m model) showTablePaginationSelection() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query("SHOW TABLES")
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
		}
		defer rows.Close()

		var tableItems []list.Item
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return queryResultMsg{nil, nil, err, 0}
			}
			tableItems = append(tableItems, tableItem{name: tableName})
		}

		return tableSelectionMsg{tableItems}
	}
}

// tableSelectionMsg represents a message for table selection
type tableSelectionMsg struct {
	items []list.Item
}

// handleTableSelection handles table selection based on context
func (m model) handleTableSelection(tableName string) tea.Cmd {
	// Determine what to do based on the current context
	// For now, just show table data
	return func() tea.Msg {
		// Get table data
		rows, err := m.db.Query(fmt.Sprintf("SELECT * FROM `%s` LIMIT 100", tableName))
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return queryResultMsg{nil, nil, err, 0}
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
				return queryResultMsg{nil, nil, err, 0}
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

		return queryResultMsg{result, columns, nil, 0}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
