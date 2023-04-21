CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  chat_id BIGINT NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  active_source TEXT,
  active_source_expire TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE source_kind AS ENUM ('unknown', 'book', 'article', 'person');
CREATE TABLE sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  kind source_kind NOT NULL DEFAULT 'unknown',
  data JSON,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quotes (
  id BIGSERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  main_source TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  text_tokens TSVECTOR GENERATED ALWAYS AS (TO_TSVECTOR('english', text || ' ' || COALESCE(main_source, ''))) STORED
);

CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE outputs (
  id BIGSERIAL PRIMARY KEY,
  chat_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users (id),
  title TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quotes_tags (
  tag BIGINT REFERENCES tags (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (tag, quote)
);

CREATE TABLE quotes_sources (
  source BIGINT REFERENCES sources (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (source, quote)
);

