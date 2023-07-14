
CREATE TABLE IF NOT EXISTS quotes_tags (
	library_id BIGINT NOT NULL REFERENCES libraries (id),
  tag BIGINT REFERENCES tags (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (tag, quote)
);
CREATE INDEX ON quotes_tags (library_id);
