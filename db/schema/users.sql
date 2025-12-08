/*******************
    Add usernames and passwords
 */
CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  username      TEXT NOT NULL UNIQUE,
  name          TEXT,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
  id         BIGSERIAL PRIMARY KEY,
  table_name TEXT NOT NULL,
  action     TEXT NOT NULL,          -- 'INSERT' | 'UPDATE' | 'DELETE'
  row_data   JSONB NOT NULL,
  by_user    BIGINT REFERENCES users(id),
  at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
