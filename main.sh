#!/bin/bash

# Check for gum
if ! command -v gum &> /dev/null; then
  echo "gum is not installed. Please install gum: https://github.com/charmbracelet/gum"
  exit 1
fi

# Prompt for connection details
HOST=$(gum input --placeholder "localhost" --prompt "🏠 Host: " --width 40)
PORT=$(gum input --placeholder "3306" --prompt "🔌 Port: " --width 10)
USER=$(gum input --placeholder "root" --prompt "👤 Username: " --width 30)
PASS=$(gum input --password --placeholder "password" --prompt "🔒 Password: " --width 30)
DB=$(gum input --placeholder "database_name" --prompt "🗄️  Database: " --width 30)

# Set defaults for empty values
HOST=${HOST:-localhost}
PORT=${PORT:-3306}
USER=${USER:-root}
DB=${DB:-database_name}

# Confirm connection
gum style --foreground 212 --border-foreground 212 --border double --align center --width 50 --padding "1 2" "Ready to connect to MySQL?"
if ! gum confirm "Connect to $USER@$HOST:$PORT/$DB?"; then
  gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "Cancelled."
  exit 0
fi

# Main menu loop
while true; do
  # Show menu options
  gum style --foreground 212 --border-foreground 212 --border double --align center --width 50 --padding "1 2" "MySQL Database Interface"
  
  CHOICE=$(gum choose \
    "📋 View Tables" \
    "🔍 Run Custom Query" \
    "📊 Show Table Data" \
    "📋 Copy Table Structure" \
    "📄 Scroll Through Results" \
    "❌ Exit")
  
  case $CHOICE in
    "📋 View Tables")
      # Show all tables in the database
      OUTPUT=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SHOW TABLES;" --batch --raw --skip-column-names --silent 2>&1)
      STATUS=$?
      
      if [[ $STATUS -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $OUTPUT"
      else
        if [[ -z "$OUTPUT" ]]; then
          gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No tables found in database"
        else
          echo "Tables in database '$DB':" | gum style --foreground 10 --border-foreground 10 --border double --align center --width 50
          echo "$OUTPUT" | gum table --columns "Table Name"
        fi
      fi
      ;;
      
    "🔍 Run Custom Query")
      # Run a custom SQL query
      QUERY=$(gum input --placeholder "Enter your SQL query..." --prompt "💬 SQL: " --width 80)
      if [[ -z "$QUERY" ]]; then
        gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "No query entered."
        continue
      fi
      
      OUTPUT=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "$QUERY" --batch --raw --skip-column-names --silent 2>&1)
      STATUS=$?
      
      if [[ $STATUS -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $OUTPUT"
      else
        if [[ -z "$OUTPUT" ]]; then
          gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "✅ Query executed successfully (no results)"
        else
          echo "$OUTPUT" | sed 's/\t/\t/g' | gum table --separator $'\t'
          gum style --foreground 10 --border-foreground 10 --border double --align center --width 50 "✅ Query executed successfully!"
        fi
      fi
      ;;
      
    "📊 Show Table Data")
      # Get table name and show its data
      TABLES=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SHOW TABLES;" --batch --raw --skip-column-names --silent 2>&1)
      if [[ $? -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $TABLES"
        continue
      fi
      
      if [[ -z "$TABLES" ]]; then
        gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No tables found in database"
        continue
      fi
      
      TABLE_NAME=$(echo "$TABLES" | gum choose)
      if [[ -z "$TABLE_NAME" ]]; then
        continue
      fi
      
      # Get column names for headers
      COLUMNS=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "DESCRIBE \`$TABLE_NAME\`;" --batch --raw --skip-column-names --silent | cut -f1)
      HEADERS=$(echo "$COLUMNS" | tr '\n' '\t')
      
      # Get table data (limit to 100 rows for display)
      OUTPUT=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SELECT * FROM \`$TABLE_NAME\` LIMIT 100;" --batch --raw --skip-column-names --silent 2>&1)
      STATUS=$?
      
      if [[ $STATUS -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $OUTPUT"
      else
        if [[ -z "$OUTPUT" ]]; then
          gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No data found in table '$TABLE_NAME'"
        else
          echo "Data from table '$TABLE_NAME' (showing first 100 rows):" | gum style --foreground 10 --border-foreground 10 --border double --align center --width 50
          # Format output properly for gum table
          echo "$OUTPUT" | sed 's/\t/\t/g' | gum table --columns "$HEADERS" --separator $'\t'
        fi
      fi
      ;;
      
    "📋 Copy Table Structure")
      # Get table name and show its structure
      TABLES=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SHOW TABLES;" --batch --raw --skip-column-names --silent 2>&1)
      if [[ $? -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $TABLES"
        continue
      fi
      
      if [[ -z "$TABLES" ]]; then
        gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No tables found in database"
        continue
      fi
      
      TABLE_NAME=$(echo "$TABLES" | gum choose)
      if [[ -z "$TABLE_NAME" ]]; then
        continue
      fi
      
      # Get table structure
      OUTPUT=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "DESCRIBE \`$TABLE_NAME\`;" --batch --raw --skip-column-names --silent 2>&1)
      STATUS=$?
      
      if [[ $STATUS -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $OUTPUT"
      else
        echo "Structure of table '$TABLE_NAME':" | gum style --foreground 10 --border-foreground 10 --border double --align center --width 50
        echo "$OUTPUT" | sed 's/\t/\t/g' | gum table --columns "Field,Type,Null,Key,Default,Extra" --separator $'\t'
        
        # Copy to clipboard if available
        if command -v clip.exe &> /dev/null; then
          echo "$OUTPUT" | clip.exe
          gum style --foreground 10 --border-foreground 10 --border double --align center --width 50 "📋 Structure copied to clipboard!"
        elif command -v pbcopy &> /dev/null; then
          echo "$OUTPUT" | pbcopy
          gum style --foreground 10 --border-foreground 10 --border double --align center --width 50 "📋 Structure copied to clipboard!"
        elif command -v xclip &> /dev/null; then
          echo "$OUTPUT" | xclip -selection clipboard
          gum style --foreground 10 --border-foreground 10 --border double --align center --width 50 "📋 Structure copied to clipboard!"
        fi
      fi
      ;;
      
    "📄 Scroll Through Results")
      # Get table name and show data with pagination
      TABLES=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SHOW TABLES;" --batch --raw --skip-column-names --silent 2>&1)
      if [[ $? -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $TABLES"
        continue
      fi
      
      if [[ -z "$TABLES" ]]; then
        gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No tables found in database"
        continue
      fi
      
      TABLE_NAME=$(echo "$TABLES" | gum choose)
      if [[ -z "$TABLE_NAME" ]]; then
        continue
      fi
      
      # Get total row count
      TOTAL_ROWS=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SELECT COUNT(*) FROM \`$TABLE_NAME\`;" --batch --raw --skip-column-names --silent 2>&1)
      if [[ $? -ne 0 ]]; then
        gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $TOTAL_ROWS"
        continue
      fi
      
      gum style --foreground 10 --border-foreground 10 --border double --align center --width 50 "📊 Table '$TABLE_NAME' has $TOTAL_ROWS rows"
      
      # Get column names for headers
      COLUMNS=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "DESCRIBE \`$TABLE_NAME\`;" --batch --raw --skip-column-names --silent | cut -f1)
      HEADERS=$(echo "$COLUMNS" | tr '\n' '\t')
      
      # Show data with pagination
      OFFSET=0
      LIMIT=20
      
      while true; do
        OUTPUT=$(mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASS" "$DB" -e "SELECT * FROM \`$TABLE_NAME\` LIMIT $LIMIT OFFSET $OFFSET;" --batch --raw --skip-column-names --silent 2>&1)
        STATUS=$?
        
        if [[ $STATUS -ne 0 ]]; then
          gum style --foreground 1 --border-foreground 1 --border double --align center --width 50 "❌ Error: $OUTPUT"
          break
        fi
        
        if [[ -z "$OUTPUT" ]]; then
          gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "📭 No more data to show"
          break
        fi
        
        echo "Rows $((OFFSET + 1))-$((OFFSET + $(echo "$OUTPUT" | wc -l))) of $TOTAL_ROWS:" | gum style --foreground 10 --border-foreground 10 --border double --align center --width 50
        echo "$OUTPUT" | sed 's/\t/\t/g' | gum table --columns "$HEADERS" --separator $'\t'
        
        NAV_CHOICE=$(gum choose "⬅️ Previous" "➡️ Next" "🔄 Refresh" "🔙 Back to Menu")
        
        case $NAV_CHOICE in
          "⬅️ Previous")
            OFFSET=$((OFFSET - LIMIT))
            if [[ $OFFSET -lt 0 ]]; then
              OFFSET=0
              gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "Already at the beginning"
            fi
            ;;
          "➡️ Next")
            OFFSET=$((OFFSET + LIMIT))
            if [[ $OFFSET -ge $TOTAL_ROWS ]]; then
              OFFSET=$((TOTAL_ROWS - LIMIT))
              if [[ $OFFSET -lt 0 ]]; then
                OFFSET=0
              fi
              gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "Already at the end"
            fi
            ;;
          "🔄 Refresh")
            # Stay on current page
            ;;
          "🔙 Back to Menu")
            break
            ;;
        esac
      done
      ;;
      
    "❌ Exit")
      gum style --foreground 212 --border-foreground 212 --border double --align center --width 50 "Goodbye!"
      exit 0
      ;;
  esac
  
  # Pause before showing menu again
  gum style --foreground 3 --border-foreground 3 --border double --align center --width 50 "Press Enter to continue..."
  read -r
done