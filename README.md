# 🐬 MySQL CLI Client

A beautiful, modern terminal-based MySQL client built with Go and Bubble Tea. Features an intuitive interface with advanced data management capabilities.

## ✨ Features

### 🎯 Core Functionality
- **Interactive SQL Query Execution** - Execute queries with real-time feedback
- **Database Connection Management** - Easy setup with connection form
- **Query History** - Navigate through previous queries with Ctrl+H
- **Tabular Results** - Beautiful table display with sorting capabilities
- **Error Handling** - Clear error messages and status indicators

### 📊 Data Management & Export
- **Multiple Copy Formats**:
  - 📋 CSV format for spreadsheet applications
  - 📊 Table format (Markdown) for documentation
  - 📄 JSON format for API integration
  - 💾 Direct file export with timestamped filenames
  - 📈 Query statistics export

### 🎨 Visual Enhancements
- **Modern UI Design** - Professional terminal interface with color coding
- **Query Statistics** - Execution time, row count, and timestamp tracking
- **Table Information** - Detailed metadata about query results
- **Status Indicators** - Real-time feedback with emojis and colors
- **Responsive Layout** - Adapts to terminal window size

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher
- MySQL server running

### Installation
```bash
git clone <repository-url>
cd boba
go mod tidy
go run main.go
```

### Usage
1. **Connect to Database**: Enter your MySQL connection details
2. **Execute Queries**: Type SQL queries and press Enter
3. **Copy Data**: Press Ctrl+D to access copy/export options
4. **View Statistics**: Press Ctrl+S to toggle query statistics
5. **Table Info**: Press Ctrl+I to toggle table information

## ⌨️ Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Execute SQL query |
| `Ctrl+C` | Quit application |
| `Esc` | Clear input field |
| `?` | Show help |
| `Ctrl+H` | Show query history |
| `Ctrl+D` | Copy/export data |
| `Ctrl+S` | Toggle query statistics |
| `Ctrl+I` | Toggle table information |
| `↑/↓` | Navigate query history |

## 📊 Copy & Export Options

When you have query results, press `Ctrl+D` to access:

1. **📋 Copy as CSV** - Export data in CSV format
2. **📊 Copy as Table** - Export as formatted Markdown table
3. **📄 Copy as JSON** - Export data in JSON format
4. **💾 Export to File** - Save data to timestamped CSV file
5. **📈 Copy Statistics** - Copy query execution statistics

## 🎨 Visual Features

### Color Scheme
- **Purple** (#7C3AED) - Primary actions and headers
- **Green** (#10B981) - Success states and secondary elements
- **Amber** (#F59E0B) - Warnings and accents
- **Red** (#EF4444) - Errors and critical states
- **Blue** (#3B82F6) - Information and statistics

### Status Indicators
- 🟢 Connected and ready
- 🔴 Error state
- ⏳ Loading/executing
- ✅ Success
- ⚠️ Warning

## 🔧 Technical Details

### Dependencies
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/go-sql-driver/mysql` - MySQL driver
- Various Bubble Tea components (table, list, textinput, etc.)

### Architecture
- **Model-View-Update** pattern with Bubble Tea
- **Modular design** with separate components for different views
- **State management** for UI state and data persistence
- **Error handling** with graceful degradation

## 🚀 Future Enhancements

- [ ] Clipboard integration for copy operations
- [ ] Query templates and snippets
- [ ] Database schema browser
- [ ] Query optimization suggestions
- [ ] Multiple database connections
- [ ] Custom themes and styling
- [ ] Export to additional formats (Excel, XML, etc.)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details. 