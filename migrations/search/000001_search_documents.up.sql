CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE search_documents (
    product_id uuid PRIMARY KEY,
    name text NOT NULL CHECK (char_length(name) BETWEEN 2 AND 160),
    description text NOT NULL CHECK (char_length(description) <= 5000),
    category_slug text NOT NULL CHECK (category_slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    updated_at timestamptz NOT NULL
);

CREATE INDEX search_documents_text_idx ON search_documents
USING gin ((name || ' ' || description) gin_trgm_ops);
CREATE INDEX search_documents_updated_idx ON search_documents (updated_at DESC, product_id DESC);

CREATE TABLE processed_search_events (
    event_id uuid PRIMARY KEY,
    processed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX processed_search_events_processed_idx ON processed_search_events (processed_at);
