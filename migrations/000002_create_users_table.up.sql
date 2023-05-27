CREATE TYPE user_state AS ENUM ('normal', 'editingSource');
CREATE TABLE IF NOT EXISTS users (
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
