# 🐬 MySQL CLI Client

A beautiful, interactive terminal-based MySQL client built with Go and Bubble Tea.

## Features

- **Interactive SQL Query Execution**: Execute SQL queries with real-time feedback
- **Query History**: Browse and reuse previous queries with `Ctrl+H`
- **Help System**: Access comprehensive help with `?` key
- **Responsive Layout**: Adapts to terminal window size
- **Error Handling**: Clear error messages and status updates
- **Tabular Results**: Beautiful table display for query results
- **Keyboard Navigation**: Full keyboard support for all operations
- **Environment Configuration**: Easy setup with `.env` file

## Interactive Features

### Key Bindings
- `Enter` - Execute SQL query
- `Ctrl+C` - Quit application
- `Esc` - Clear input field
- `?` - Show/hide help
- `Ctrl+H` - Show query history
- `↑/↓` - Navigate through query history

### Query History
- Automatically saves executed queries
- Browse history with `Ctrl+H`
- Select and reuse previous queries
- Navigate through history with arrow keys

### Help System
- Press `?` to access the help screen
- Comprehensive key binding documentation
- Feature overview
- Press `Esc` to return to main interface

## Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Configure your database connection (see Configuration section)
4. Build the application:
   ```bash
   go build -o boba.exe main.go
   ```

## Configuration

### Using .env file (Recommended)

1. Copy the sample environment file:
   ```bash
   cp .env_sample .env
   ```

2. Edit the `.env` file with your database credentials:
   ```bash
   # Database Host (default: localhost)
   DB_HOST=localhost
   
   # Database Port (default: 3306)
   DB_PORT=3306
   
   # Database Username (default: root)
   DB_USER=root
   
   # Database Password (default: empty)
   DB_PASS=your_password_here
   
   # Database Name (default: empty)
   DB_NAME=your_database_name
   ```

### Using Environment Variables

Alternatively, you can set environment variables directly:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASS=your_password
export DB_NAME=your_database
```

### Default Values

If no environment variables are set, the application uses these defaults:
- Host: localhost
- Port: 3306
- User: root
- Password: (empty)
- Database: (empty)

## Usage

1. Start the application:
   ```bash
   ./boba.exe
   ```

2. Enter SQL queries in the input field
3. Press `Enter` to execute
4. Use `?` for help or `Ctrl+H` for history
5. Press `Ctrl+C` to quit

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [MySQL Driver](https://github.com/go-sql-driver/mysql) - Database connectivity
- [Godotenv](https://github.com/joho/godotenv) - Environment variable loading

## Screenshots

The application features:
- Clean, modern interface with emoji icons
- Color-coded status messages
- Responsive table layouts
- Interactive help and history screens
- Smooth keyboard navigation

## Contributing

Feel free to submit issues and enhancement requests! 