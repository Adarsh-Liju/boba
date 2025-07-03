package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

type dbCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
}

type queryRequest struct {
	Credentials dbCredentials `json:"credentials"`
	Query       string        `json:"query"`
}

func connectToDatabase(dbCredentials dbCredentials) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbCredentials.Username, dbCredentials.Password, dbCredentials.Host, dbCredentials.Port, dbCredentials.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func setupRouter() *gin.Engine {
	// Create a new Gin router
	r := gin.Default()

	r.StaticFile("/", "./index.html")

	r.POST("/login", func(c *gin.Context) {
		var dbCredentials dbCredentials
		if err := c.ShouldBindJSON(&dbCredentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db, err := connectToDatabase(dbCredentials)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database: " + err.Error()})
			return
		}
		defer db.Close()
		c.JSON(http.StatusOK, gin.H{"message": "Database connected successfully"})
	})

	r.POST("/execute-query", func(c *gin.Context) {
		var req queryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate query is not empty
		if req.Query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
			return
		}

		db, err := connectToDatabase(req.Credentials)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database: " + err.Error()})
			return
		}
		defer db.Close()

		rows, err := db.Query(req.Query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		results := []map[string]any{}
		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			row := make(map[string]any)
			for i, col := range columns {
				val := values[i]
				if val == nil {
					row[col] = nil
				} else {
					// Handle different data types properly
					switch v := val.(type) {
					case []byte:
						// Handle BLOB/TEXT fields
						row[col] = string(v)
					case int64:
						row[col] = v
					case int32:
						row[col] = v
					case int:
						row[col] = v
					case float64:
						row[col] = v
					case float32:
						row[col] = v
					case bool:
						row[col] = v
					case string:
						row[col] = v
					default:
						// For any other type, convert to string safely
						row[col] = fmt.Sprintf("%v", v)
					}
				}
			}
			results = append(results, row)
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"count":   len(results),
		})
	})

	return r
}

func main() {
	r := setupRouter()
	log.Println("Server starting on :8080")
	r.Run(":8080")
}
