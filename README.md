# boba - Interactive MySQL CLI with Bubble Tea

A modern, interactive MySQL command-line interface built with Go and Bubble Tea for a beautiful terminal experience.

## Features

- 🔗 Simple database connection
- 💬 Interactive SQL query execution
- 📚 Browse and switch between databases
- 📋 View and explore tables
- 🎨 Beautiful terminal UI with Bubble Tea
- 📊 Clean table output with navigation
- 🚀 Easy to use and understand

## Installation

1. Make sure you have Go installed
2. Clone this repository
3. Install dependencies:
```bash
go mod tidy
```

## Usage

Run the application:
```bash
go run main.go
```

The program will prompt you for:
- **Host** (defaults to localhost)
- **User** (defaults to root)
- **Password**
- **Database name**

After connecting, you'll see an interactive menu with options:
- **Execute Query** - Type and run SQL queries
- **View Databases** - Browse all available databases
- **View Tables** - See tables in current database
- **Exit** - Quit the application

## Navigation

- **↑↓** - Navigate through menus and lists
- **Enter** - Select an option or execute a query
- **Escape** - Go back to previous menu
- **Ctrl+C** - Quit the application

## Example

```
🔗 MySQL Connection

Host: localhost
User: root
Database: testdb
Password: ********

Press Enter to connect, Ctrl+C to quit
```

After connecting:
```
✅ Connected to MySQL at localhost/testdb

📋 Main Menu
─────────────

> Execute Query
  View Databases
  View Tables
  Exit

Use ↑↓ to navigate, Enter to select, Ctrl+C to quit
```

## Dependencies

- `github.com/go-sql-driver/mysql` - MySQL driver for Go
- `github.com/charmbracelet/bubbletea` - Terminal UI framework

## Building

To build the executable:
```bash
go build -o boba main.go
```

Then run:
```bash
./boba
```

## Features in Detail

### Execute Query
- Type SQL queries and press Enter to execute
- Results are displayed in a clean table format
- Shows up to 10 rows with pagination info
- Press Escape to return to main menu

### View Databases
- Browse all available databases on the server
- Select a database to switch to it
- Automatically reconnects with the new database

### View Tables
- See all tables in the current database
- Select a table to view its data (first 10 rows)
- Automatically generates and executes `SELECT * FROM table LIMIT 10`

## TODO

1. Add more colors and cute stuff
2. Improve UI
