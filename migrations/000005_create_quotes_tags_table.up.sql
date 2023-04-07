
CREATE TABLE IF NOT EXISTS quotes_tags (
  tag BIGINT REFERENCES tags (id),
  quote BIGINT REFERENCES quotes (id),
  PRIMARY KEY (tag, quote)
);