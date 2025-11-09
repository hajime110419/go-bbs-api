package repository

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// CreateTable ensures the necessary database schema (the 'posts' table) exists.
// It accepts a database connection pool and can be used for both the main application
// and for setting up test databases.
func CreateTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
		"id" TEXT NOT NULL PRIMARY KEY,
		"title" TEXT,
		"content" TEXT
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		// Since this is a critical startup step, we terminate if it fails.
		log.Fatalf("Failed to create table: %v", err)
	}
}

// InitDB now focuses solely on opening the main application's database file
// and then calls CreateTable to set up the schema.
// It returns the database connection so it can be used throughout the application.
func InitDB() *sql.DB {
	var err error
	// The driver name "sqlite" is registered by the blank import of modernc.org/sqlite.
	db, err := sql.Open("sqlite", "./bulletinboard.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Call the separated function to create the table schema.
	CreateTable(db)

	return db
}
