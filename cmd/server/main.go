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
	_ = godotenv.Load()

	// DB connection (fallback to SQLite if DATABASE_URL is missing)
	dsn := os.Getenv("DATABASE_URL")
	driver := "postgres"

	if dsn == "" {
		driver = "sqlite3"
		dsn = "file:knowledge.db?cache=shared&mode=rwc"
		fmt.Println("DATABASE_URL not set. Using local SQLite DB.")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	defer db.Close()

	// Decide which LLM client to use
	var llmClient llm.LLM
	if os.Getenv("USE_MOCK_LLM") == "true" || os.Getenv("OPENAI_API_KEY") == "" {
		llmClient = llm.NewMockClient()
		fmt.Println("Using Mock LLM Client")
	} else {
		openaiClient := llm.NewOpenAIClient()
		llmClient = llm.NewResilientClient(openaiClient, llm.NewMockClient())
		fmt.Println("Using OpenAI LLM Client (with automatic mock fallback)")
	}

	// Create server
	s := server.New(db, llmClient)

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
