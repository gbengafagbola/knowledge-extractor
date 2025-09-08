
# LLM Knowledge Extractor

This project is a prototype that ingests unstructured text, uses an LLM (OpenAI or a mock fallback) to analyze it, and returns structured data along with persistence to a database.

---

## Features

* **Analyze Text** (`POST /analyze`)

  * Generates a **1–2 sentence summary**.
  * Extracts:

    * `title`
    * `topics` (3 key topics)
    * `sentiment` (positive, neutral, negative)
    * `keywords` (3 most frequent nouns — implemented locally, not via LLM)
    * `confidence` score (simple heuristic).
  * Stores results in Postgres (Supabase) or SQLite fallback.

* **Search Analyses** (`GET /search?topic=xyz`)

  * Returns all stored analyses with matching topic/keyword.

* **Resilient LLM Client**

  * Uses **OpenAI** if available.
  * Automatically **falls back to a mock client** per request if the OpenAI API call fails.
  * Can be forced into mock-only mode via `USE_MOCK_LLM=true`.

* Handles edge cases:

  * Empty input
  * LLM API failure

---

## Tech Stack

* **Golang** (API server)
* **Postgres (Supabase)** — primary persistence
* **SQLite** — local fallback database
* **OpenAI GPT API** — text analysis (with mock fallback)

---

## Setup & Run

### 1. Clone the repo

```bash
git clone https://github.com/gbengafagbola/knowledge-extractor.git
cd knowledge-extractor
```

### 2. Configure environment

Create a `.env` file in the project root:

```env
# Primary database (Postgres / Supabase)
DATABASE_URL=postgres://<user>:<password>@<host>:5432/postgres?sslmode=require

# OpenAI API key
OPENAI_API_KEY=sk-...

# Force mock mode (true/false)
USE_MOCK_LLM=false

# Server port
PORT=8080
```

If `DATABASE_URL` is not set, the app falls back to SQLite (`knowledge.db`).

### 3. Run the server

```bash
go run ./cmd/server
```

Expected output (if OpenAI is configured):

```
Using OpenAI LLM Client (with automatic mock fallback)
Server running on port 8080
```

Or (if mock mode):

```
 Using Mock LLM Client
Server running on port 8080
```

### 4. Example Requests

#### Analyze text

```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"text": "Summarize quantum computing in simple terms."}'
```

#### Search analyses

```bash
curl "http://localhost:8080/search?topic=quantum"
```

---

## Design Choices

* **Go** was chosen for performance, explicit error handling, and strong typing, even though it requires more boilerplate than Python.
* **Postgres (Supabase)** ensures durability and cloud compatibility, while **SQLite** provides a lightweight local fallback for quick testing.
* The **LLM abstraction** (`internal/llm/LLM`) decouples the API from the specific model (OpenAI or Mock), making testing and fallback straightforward.
* The **ResilientClient** pattern ensures robustness: every request tries OpenAI first, and if it fails, the system transparently falls back to mock results.

---

## Trade-offs

* Timeboxing limited the implementation of a full test suite; only manual testing with curl is provided.
* No authentication or user management was added.
* The confidence score is a naive static heuristic.
* API responses are simple JSON without pagination or advanced search.

---

## Next Steps (if extended)

* Add **unit and integration tests**.
* Containerize with **Docker** for easy deployment.  (couldn't complete in given time window)
* Add a minimal **web UI** to submit text and browse results (couldn't complete in given time window).
* Improve keyword extraction (currently based on basic noun frequency).


# knowledge-extractor
