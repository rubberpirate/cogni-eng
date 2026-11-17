-- +goose Up
CREATE TABLE admission_code (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,
 code_hash text NOT NULL UNIQUE,label text NOT NULL,expires_at timestamptz NOT NULL,max_uses integer NOT NULL CHECK(max_uses BETWEEN 1 AND 1000),
 used integer NOT NULL DEFAULT 0 CHECK(used>=0),revoked_at timestamptz,created_by text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id)
);
CREATE TABLE admission_request (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,person_id text NOT NULL REFERENCES person(id),
 code_id text NOT NULL REFERENCES admission_code(id),state text NOT NULL DEFAULT 'unverified' CHECK(state IN ('unverified','pending','placed','rejected')),
 reason text NOT NULL DEFAULT '',version integer NOT NULL DEFAULT 1,created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),UNIQUE(institution_id,person_id)
);
CREATE TABLE policy_acknowledgment (
 id text PRIMARY KEY,person_id text NOT NULL REFERENCES person(id),policy text NOT NULL,version text NOT NULL,acknowledged_at timestamptz NOT NULL DEFAULT now(),UNIQUE(person_id,policy,version)
);
-- +goose Down
DROP TABLE policy_acknowledgment,admission_request,admission_code;
