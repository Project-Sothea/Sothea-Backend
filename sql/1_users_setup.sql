/*******************
    Drop the tables
********************/
DROP TABLE IF EXISTS users;

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

-- Insert the users
INSERT INTO users (username, name, password_hash)
VALUES ('admin', 'admin', '$2a$10$Y/FMxVZsuiPcVutd0O3kfuWEyWyqZUb4HC5yXF7.xVxNCt9GUfOPe');
INSERT INTO users (username, name, password_hash)
VALUES ('user', 'user', '$2a$10$cqFdRzsZVwGF6fI4N.YEYOOEZL7B/93RmEZRkVPVHHqyBMfNwy48i');

--------------------------------------------------------------------------------
-- GLOBAL AUDIT HELPERS (used by ALL other tables; keep them in this file)
--------------------------------------------------------------------------------

-- 1) Fills created_by/updated_by + timestamps using a per-transaction GUC
--    (set by  code: SET LOCAL sothea.user_id = $1)
CREATE OR REPLACE FUNCTION set_audit_fields() RETURNS TRIGGER AS $$
DECLARE v_user_id BIGINT;
BEGIN
  BEGIN
    v_user_id := NULLIF(current_setting('sothea.user_id', true), '')::BIGINT;
  EXCEPTION WHEN others THEN
    v_user_id := NULL;
  END;

  IF TG_OP = 'INSERT' THEN
    IF NEW.created_at IS NULL THEN NEW.created_at := now(); END IF;
    IF NEW.updated_at IS NULL THEN NEW.updated_at := now(); END IF;
    IF NEW.created_by IS NULL THEN NEW.created_by := v_user_id; END IF;
    IF NEW.updated_by IS NULL THEN NEW.updated_by := v_user_id; END IF;

  ELSIF TG_OP = 'UPDATE' THEN
    NEW.updated_at := now();
    IF v_user_id IS NOT NULL THEN NEW.updated_by := v_user_id; END IF;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2) Convenience helper: add audit columns + attach the above trigger to any table
CREATE OR REPLACE FUNCTION add_audit(p_table regclass) RETURNS void AS $$
DECLARE
  relname text := (SELECT relname FROM pg_class WHERE oid = p_table);
  trgname text := 'trg_' || relname || '_audit';
BEGIN
  EXECUTE format('ALTER TABLE %s
    ADD COLUMN IF NOT EXISTS created_by BIGINT REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS updated_by BIGINT REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();',
    p_table
  );

  -- Replace any existing audit trigger for idempotency
  IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = trgname AND NOT tgisinternal) THEN
    EXECUTE format('DROP TRIGGER %I ON %s;', trgname, p_table);
  END IF;

  EXECUTE format('CREATE TRIGGER %I
    BEFORE INSERT OR UPDATE ON %s
    FOR EACH ROW EXECUTE FUNCTION set_audit_fields();',
    trgname, p_table
  );
END;
$$ LANGUAGE plpgsql;

-- (Optional) Structured change log for inserts/updates/deletes; attach later if wanted.
CREATE TABLE IF NOT EXISTS audit_log (
  id         BIGSERIAL PRIMARY KEY,
  table_name TEXT NOT NULL,
  action     TEXT NOT NULL,          -- 'INSERT' | 'UPDATE' | 'DELETE'
  row_data   JSONB NOT NULL,
  by_user    BIGINT REFERENCES users(id),
  at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION audit_row() RETURNS TRIGGER AS $$
DECLARE v_uid BIGINT := NULLIF(current_setting('sothea.user_id', true), '')::BIGINT;
BEGIN
  IF TG_OP = 'DELETE' THEN
    INSERT INTO audit_log(table_name, action, row_data, by_user)
    VALUES (TG_TABLE_NAME, 'DELETE', to_jsonb(OLD), v_uid);
    RETURN OLD;
  ELSIF TG_OP = 'INSERT' THEN
    INSERT INTO audit_log(table_name, action, row_data, by_user)
    VALUES (TG_TABLE_NAME, 'INSERT', to_jsonb(NEW), v_uid);
    RETURN NEW;
  ELSE
    INSERT INTO audit_log(table_name, action, row_data, by_user)
    VALUES (TG_TABLE_NAME, 'UPDATE', jsonb_build_object('old', to_jsonb(OLD), 'new', to_jsonb(NEW)), v_uid);
    RETURN NEW;
  END IF;
END; $$ LANGUAGE plpgsql;
