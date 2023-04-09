
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY,
  chat_id BIGINT NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  active_source TEXT,
  active_source_expire TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
