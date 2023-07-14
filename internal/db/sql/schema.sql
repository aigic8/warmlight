CREATE TYPE user_state AS ENUM ('normal', 'editingSource', 'changingLibrary', 'confirmingLibraryChange');
CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  chat_id BIGINT NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  state user_state NOT NULL DEFAULT 'normal',
  state_data JSON,
  active_source TEXT,
  active_source_expire TIMESTAMPTZ,
	library_id BIGINT NOT NULL REFERENCES libraries (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE libraries (
	id BIGSERIAL PRIMARY KEY,
	owner_id BIGINT NOT NULL,
	token UUID,
	token_expires_on TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE source_kind AS ENUM ('unknown', 'book', 'article', 'person');
CREATE TABLE sources (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  library_id BIGINT NOT NULL REFERENCES libraries (id),
  kind source_kind NOT NULL DEFAULT 'unknown',
  data JSON,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE quotes (
  id BIGSERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  library_id BIGINT NOT NULL REFERENCES libraries (id),
  main_source TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  text_tokens TSVECTOR GENERATED ALWAYS AS (TO_TSVECTOR('english', text || ' ' || COALESCE(main_source, ''))) STORED
);

CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  library_id BIGINT NOT NULL REFERENCES libraries (id),
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
	library_id BIGINT NOT NULL REFERENCES libraries (id),
  tag BIGINT REFERENCES tags (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (tag, quote)
);

CREATE TABLE quotes_sources (
	library_id BIGINT NOT NULL REFERENCES libraries (id),
  source BIGINT REFERENCES sources (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (source, quote)
);

