CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE apps (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name            TEXT        NOT NULL,
  public_key      TEXT        NOT NULL UNIQUE,
  allowed_origins TEXT[]      NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at     TIMESTAMPTZ
);

CREATE INDEX apps_active_key_idx
  ON apps (public_key)
  WHERE archived_at IS NULL;