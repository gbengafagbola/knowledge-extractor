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
);
