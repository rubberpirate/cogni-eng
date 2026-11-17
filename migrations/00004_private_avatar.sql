-- +goose Up
CREATE TABLE avatar_asset (
 id text PRIMARY KEY,person_id text NOT NULL REFERENCES person(id),object_key text NOT NULL UNIQUE,
 state text NOT NULL CHECK(state IN ('staged','active','retired','purged')),
 bytes integer NOT NULL CHECK(bytes>0),sha256 text NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),purge_after timestamptz,
 purged_at timestamptz,last_error text NOT NULL DEFAULT ''
);
ALTER TABLE person ADD COLUMN avatar_asset_id text REFERENCES avatar_asset(id);
CREATE INDEX avatar_purge ON avatar_asset(purge_after) WHERE state IN ('staged','retired');
-- +goose Down
ALTER TABLE person DROP COLUMN avatar_asset_id;
DROP TABLE avatar_asset;
