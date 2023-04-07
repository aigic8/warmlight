
CREATE TABLE IF NOT EXISTS quotes_sources (
  source BIGINT REFERENCES sources (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (source, quote)
);