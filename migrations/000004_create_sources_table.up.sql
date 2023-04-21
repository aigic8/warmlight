
CREATE TYPE source_type AS ENUM ('unknown', 'book', 'website', 'person');

CREATE TABLE IF NOT EXISTS sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  kind source_type NOT NULL DEFAULT 'unknown',
  data JSON,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ON sources (user_id, name);