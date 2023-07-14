
CREATE TABLE IF NOT EXISTS quotes (
  id BIGSERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  library_id BIGINT NOT NULL REFERENCES libraries (id),
  main_source TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  text_tokens TSVECTOR GENERATED ALWAYS AS (TO_TSVECTOR('english', text || ' ' || COALESCE(main_source, ''))) STORED
);

CREATE UNIQUE INDEX ON quotes (library_id, text);
CREATE INDEX ON quotes USING GIN (text_tokens);
