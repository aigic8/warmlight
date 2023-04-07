
CREATE TABLE IF NOT EXISTS quotes (
  id BIGSERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  main_source TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  text_tokens TSVECTOR GENERATED ALWAYS AS (TO_TSVECTOR('english', text || ' ' || COALESCE(main_source, ''))) STORED
);

CREATE UNIQUE INDEX ON quotes (user_id, text);
CREATE INDEX ON quotes USING GIN (text_tokens);