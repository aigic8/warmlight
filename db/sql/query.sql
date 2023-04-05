-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, chat_id, firstname) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateSource :one
INSERT INTO sources (user_id, name) VALUES ($1, $2) RETURNING *;

-- name: GetSource :one
SELECT * FROM sources WHERE user_id = $1 AND name = $2;

-- name: SetActiveSource :one
UPDATE users SET
  active_source = $2,
  active_source_expire = $3
WHERE id = $1 RETURNING *;

-- name: DeactivateExpiredSources :many
UPDATE users SET
  active_source = NULL,
  active_source_expire = NULL
WHERE active_source_expire <= NOW() RETURNING *;

-- name: DeactivateSource :one
UPDATE users SET
  active_source = NULL,
  active_source_expire = NULL
WHERE id = $1 RETURNING *;

-- name: GetOutputs :many
SELECT * FROM outputs WHERE user_id = $1;

-- name: GetOutput :one
SELECT * FROM outputs WHERE user_id = $1 AND title = $2;

-- name: SetOutputActive :one
UPDATE outputs SET
  is_active = TRUE
WHERE user_id = $1 AND title = $2 RETURNING *;

-- name: CreateOutput :one
INSERT INTO outputs (user_id, chat_id, title) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteOutput :exec
DELETE FROM outputs WHERE user_id = $1 AND title = $2;

-- name: CleanOutputs :exec
DELETE FROM outputs; 

-- name: CleanQuotesTags :exec
DELETE FROM quotes_tags; 

-- name: CleanQuotesSources :exec
DELETE FROM quotes_sources; 

-- name: CleanTags :exec
DELETE FROM tags; 

-- name: CleanSources :exec
DELETE FROM sources; 

-- name: CleanQuotes :exec
DELETE FROM quotes; 

-- name: CleanUsers :exec
DELETE FROM users;

-- name: CreateQuote :one
INSERT INTO quotes (user_id, text, main_source) VALUES ($1, $2, $3) RETURNING id, text, user_id, main_source, created_at, updated_at;

-- name: GetOrCreateTag :one
WITH created_id AS (
  INSERT INTO tags (user_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id
) SELECT id FROM created_id UNION ALL SELECT id FROM tags WHERE user_id = $1 AND name = $2 LIMIT 1;

-- name: GetOrCreateSource :one
WITH created_id AS (
  INSERT INTO sources (user_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id
) SELECT id FROM created_id UNION ALL SELECT id FROM sources WHERE user_id = $1 AND name = $2 LIMIT 1;

-- name: CreateQuotesTags :exec
INSERT INTO quotes_tags (quote, tag) VALUES ($1, $2);

-- name: CreateQuotesSources :exec
INSERT INTO quotes_sources (quote, source) VALUES ($1, $2);

-- name: SearchQuotes :many
SELECT id, text, main_source, user_id, created_at, updated_at FROM quotes WHERE user_id = $1 AND text_tokens @@ TO_TSQUERY('english', $2) LIMIT $3;