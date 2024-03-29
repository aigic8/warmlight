
CREATE TABLE IF NOT EXISTS libraries (
	id BIGSERIAL PRIMARY KEY,
	owner_id BIGINT NOT NULL,
	token UUID,
	token_expires_on TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX ON libraries (token);
