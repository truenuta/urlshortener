CREATE TABLE IF NOT EXISTS urls (
      uuid         UUID PRIMARY KEY,
      short_url    VARCHAR(255) NOT NULL UNIQUE,
      original_url VARCHAR(255) NOT NULL
  );