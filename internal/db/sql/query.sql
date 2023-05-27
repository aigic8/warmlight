---------- USERS -------------

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (id, chat_id, first_name, library_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: SetUserState :one
UPDATE users SET
  state = $2,
  state_data = $3
WHERE id = $1 RETURNING *;

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

--------- LIBRARIES ----------

-- name: GetLibrary :one
SELECT * FROM libraries WHERE id = $1;

-- name: CreateLibrary :one
INSERT INTO libraries (owner_id) VALUES ($1) RETURNING *;

---------- QUOTES ------------

-- name: CreateQuote :one
INSERT INTO quotes (library_id, text, main_source) VALUES ($1, $2, $3) RETURNING id, text, library_id, main_source, created_at, updated_at;

-- name: SearchQuotes :many
SELECT id, text, main_source, library_id, created_at, updated_at FROM quotes WHERE library_id = $1 AND text_tokens @@ TO_TSQUERY('english', $2) LIMIT $3;

---------- SOURCES -----------

-- name: GetSource :one
SELECT * FROM sources WHERE library_id = $1 AND name = $2;

-- name: CreateSource :one
INSERT INTO sources (library_id, name) VALUES ($1, $2) RETURNING *;

-- name: GetSourceByID :one
SELECT * FROM sources WHERE library_id = $1 AND id = $2;

-- name: GetOrCreateSource :one
WITH created_id AS (
  INSERT INTO sources (library_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id
) SELECT id FROM created_id UNION ALL SELECT id FROM sources WHERE library_id = $1 AND name = $2 LIMIT 1;

-- name: QuerySourcesAfter :many
SELECT * FROM sources WHERE library_id = $1 AND id > $2 AND name LIKE '%' || $3 || '%' ORDER BY id ASC LIMIT $4;

-- name: QuerySourcesAfterWithKind :many
SELECT * FROM sources WHERE library_id = $1 AND id > $2 AND kind = $3 AND name LIKE '%' || $4 || '%' ORDER BY id ASC LIMIT $5;

-- name: QuerySourcesBefore :many
SELECT * FROM sources WHERE library_id = $1 AND id < $2 AND name LIKE '%' || $3 || '%' ORDER BY id DESC LIMIT $4;

-- name: QuerySourcesBeforeWithKind :many
SELECT * FROM sources WHERE library_id = $1 AND id < $2 AND kind = $3 AND name LIKE '%' || $4 || '%' ORDER BY id DESC LIMIT $5;

-- name: SetSourceData :one
UPDATE sources SET kind = $1, data = $2  WHERE library_id = $3 AND id = $4 RETURNING *;

-- name: UpdateSource :one
UPDATE sources SET name = $1, kind = $2, data = $3, updated_at = NOW() WHERE library_id = $4 AND id = $5 RETURNING *;

---------- OUTPUTS -----------

-- name: CreateOutput :one
INSERT INTO outputs (user_id, chat_id, title) VALUES ($1, $2, $3) RETURNING *;

-- name: GetOutputs :many
SELECT * FROM outputs WHERE user_id = $1;

-- name: GetOutput :one
SELECT * FROM outputs WHERE user_id = $1 AND chat_id = $2;

-- name: ActivateOutput :one
UPDATE outputs SET
  is_active = TRUE
WHERE chat_id = $1 AND user_id = $2 RETURNING *;

-- name: DeactivateOutput :one
UPDATE outputs SET
  is_active = FALSE
WHERE chat_id = $1 AND user_id = $2 RETURNING *;

-- name: DeleteOutput :exec
DELETE FROM outputs WHERE user_id = $1 AND chat_id = $2;

----------- TAGS -------------

-- name: GetOrCreateTag :one
WITH created_id AS (
  INSERT INTO tags (library_id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id
) SELECT id FROM created_id UNION ALL SELECT id FROM tags WHERE library_id = $1 AND name = $2 LIMIT 1;

-------- ASSOCIATIONS ---------

-- name: CreateQuotesTags :exec
INSERT INTO quotes_tags (quote, tag) VALUES ($1, $2);

-- name: CreateQuotesSources :exec
INSERT INTO quotes_sources (quote, source) VALUES ($1, $2);

----------- DEBUG ------------

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
