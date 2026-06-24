package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"backend/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB connects to PostgreSQL, ensures the DB exists, and runs migrations
func InitDB() {
	cfg := config.ActiveConfig

	// 1. Connect to default 'postgres' database to ensure our 'flowak' DB exists
	adminConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
	
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		log.Printf("Warning: Failed to connect to default postgres DB (will try direct connection): %v", err)
	} else {
		// Try to create the database if it doesn't exist
		// Exec in PostgreSQL doesn't support parameterized DB names, but credentials are safe
		var exists bool
		err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.DBName).Scan(&exists)
		if err == nil && !exists {
			log.Printf("Database '%s' does not exist. Creating database...", cfg.DBName)
			_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
			if err != nil {
				log.Printf("Warning: Failed to create database '%s' automatically: %v", cfg.DBName, err)
			} else {
				log.Printf("Database '%s' created successfully.", cfg.DBName)
			}
		}
		adminDB.Close()
	}

	// 2. Connect to the actual target database
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open PostgreSQL connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	log.Printf("Connected to PostgreSQL database '%s' successfully.", cfg.DBName)

	// 3. Run migrations
	runMigrations()
}

func runMigrations() {
	log.Println("Running database migrations from SQL files...")
	
	// Execute schema migration
	runMigrationFromFile(DB, "01_init_schema.sql")
	
	// Execute seed data migration
	runMigrationFromFile(DB, "02_seed_data.sql")

	log.Println("Migrations executed successfully.")
}

func runMigrationFromFile(db *sql.DB, filename string) {
	// Search paths for migration folder to be CWD agnostic
	paths := []string{
		filepath.Join("db", "migrations", filename),
		filepath.Join("migrations", filename),
		filepath.Join("backend", "db", "migrations", filename),
		filepath.Join("..", "backend", "db", "migrations", filename),
	}

	var content []byte
	var err error
	var foundPath string
	
	for _, p := range paths {
		content, err = os.ReadFile(p)
		if err == nil {
			foundPath = p
			break
		}
	}

	if err != nil {
		log.Fatalf("Failed to read migration file %s in any checked paths: %v", filename, err)
	}

	log.Printf("Executing migration: %s", foundPath)
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute SQL migration %s: %v", filename, err)
	}
}
