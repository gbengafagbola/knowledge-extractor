package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gbengafagbola/knowledge-extractor/internal/llm"
	"github.com/gbengafagbola/knowledge-extractor/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	// This allows for easy configuration management across environments
	_ = godotenv.Load()

	// DATABASE STRATEGY: Graceful degradation pattern
	// Try PostgreSQL (production) first, fallback to SQLite (development)
	// This ensures the application works in both cloud and local environments
	dsn := os.Getenv("DATABASE_URL")
	driver := "postgres"
	var db *sql.DB
	var err error

	if dsn != "" {
		// Try PostgreSQL first
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			fmt.Println("Failed to open PostgreSQL connection:", err)
			fmt.Println("Falling back to SQLite...")
		} else {
			// Test the connection
			err = db.Ping()
			if err != nil {
				fmt.Println("PostgreSQL connection test failed:", err)
				fmt.Println("Falling back to SQLite...")
				db.Close() // Close the failed connection
			} else {
				fmt.Println("Connected to PostgreSQL successfully!")
			}
		}
	}

	// Use SQLite if PostgreSQL failed or wasn't configured
	if dsn == "" || err != nil {
		driver = "sqlite3"
		dsn = "file:knowledge.db?cache=shared&mode=rwc"
		fmt.Println("Using local SQLite DB.")

		db, err = sql.Open(driver, dsn)
		if err != nil {
			log.Fatal("failed to open SQLite db:", err)
		}
	}

	defer db.Close()

	// Create table if it doesn't exist (works for both PostgreSQL and SQLite)
	if err := createTableIfNotExists(db, driver); err != nil {
		log.Fatal("failed to create table:", err)
	}

	// LLM STRATEGY: Interface-based dependency injection with resilience
	// Demonstrates several design patterns:
	// 1. Strategy Pattern: Different LLM implementations
	// 2. Decorator Pattern: ResilientClient wraps OpenAI with fallback
	// 3. Dependency Inversion: Server depends on interface, not concrete types
	var llmClient llm.LLM
	if os.Getenv("USE_MOCK_LLM") == "true" || os.Getenv("OPENAI_API_KEY") == "" {
		llmClient = llm.NewMockClient()
		fmt.Println("Using Mock LLM Client")
	} else {
		openaiClient := llm.NewOpenAIClient()
		// ResilientClient implements circuit breaker pattern
		llmClient = llm.NewResilientClient(openaiClient, llm.NewMockClient())
		fmt.Println("Using OpenAI LLM Client (with automatic mock fallback)")
	}

	// Create server
	s := server.New(db, llmClient, driver)

	// Routes
	http.HandleFunc("/analyze", s.AnalyzeHandler)
	http.HandleFunc("/search", s.SearchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// createTableIfNotExists creates the analyses table for both PostgreSQL and SQLite
func createTableIfNotExists(db *sql.DB, driver string) error {
	var createTableSQL string

	if driver == "postgres" {
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS analyses (
				id TEXT PRIMARY KEY,
				raw_text TEXT NOT NULL,
				summary TEXT,
				title TEXT,
				topics TEXT[],
				sentiment TEXT,
				keywords TEXT[],
				confidence NUMERIC,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
			);`
	} else {
		// SQLite version - arrays stored as comma-separated strings
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS analyses (
				id TEXT PRIMARY KEY,
				raw_text TEXT NOT NULL,
				summary TEXT,
				title TEXT,
				topics TEXT,
				sentiment TEXT,
				keywords TEXT,
				confidence REAL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`
	}

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Println("Database table 'analyses' created/verified successfully")
	return nil
}
