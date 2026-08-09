DROP INDEX search_documents_text_idx;
CREATE INDEX search_documents_name_idx ON search_documents USING gin (name gin_trgm_ops);
CREATE INDEX search_documents_description_idx ON search_documents USING gin (description gin_trgm_ops);
