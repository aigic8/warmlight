
CREATE TABLE IF NOT EXISTS quotes_sources (
	library_id BIGINT NOT NULL REFERENCES libraries (id),
  source BIGINT REFERENCES sources (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (source, quote)
);
CREATE INDEX ON quotes_sources (library_id);
