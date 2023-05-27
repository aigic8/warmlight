
CREATE TYPE source_kind AS ENUM ('unknown', 'book', 'article', 'person');

CREATE TABLE IF NOT EXISTS sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  library_id BIGINT NOT NULL REFERENCES libraries (id),
  kind source_kind NOT NULL DEFAULT 'unknown',
  data JSON,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ON sources (library_id, name);
