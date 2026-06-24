package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB opens the SQLite database and runs migrations
func InitDB(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping SQLite database: %v", err)
	}

	log.Println("Database connection established at", dbPath)
	migrate()
}

func migrate() {
	// Enable foreign key constraints in SQLite
	_, err := DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// 1. Users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = DB.Exec(usersTable)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// 2. Projects table
	projectsTable := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		owner_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(owner_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	_, err = DB.Exec(projectsTable)
	if err != nil {
		log.Fatalf("Failed to create projects table: %v", err)
	}

	// 3. Modules table
	modulesTable := `
	CREATE TABLE IF NOT EXISTS modules (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		nodes TEXT NOT NULL DEFAULT '[]',
		edges TEXT NOT NULL DEFAULT '[]',
		schema_version INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`
	_, err = DB.Exec(modulesTable)
	if err != nil {
		log.Fatalf("Failed to create modules table: %v", err)
	}

	log.Println("Database migration completed successfully.")
}
