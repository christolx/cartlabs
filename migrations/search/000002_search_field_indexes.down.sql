DROP INDEX IF EXISTS search_documents_description_idx;
DROP INDEX IF EXISTS search_documents_name_idx;
CREATE INDEX search_documents_text_idx ON search_documents
USING gin ((name || ' ' || description) gin_trgm_ops);
