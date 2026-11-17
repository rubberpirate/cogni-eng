-- +goose Up
CREATE TABLE person (
 id text PRIMARY KEY, email text NOT NULL UNIQUE CHECK(email=lower(email)),
 name text NOT NULL CHECK(length(name) BETWEEN 1 AND 120), password_hash text NOT NULL,
 state text NOT NULL CHECK(state IN ('invited','active','suspended','archived')),
 verified_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(),
 state_reason text NOT NULL DEFAULT '', mfa_secret bytea, mfa_enabled boolean NOT NULL DEFAULT false,
 last_totp_step bigint NOT NULL DEFAULT 0
);
CREATE TABLE institution (
 id text PRIMARY KEY, name text NOT NULL CHECK(length(name) BETWEEN 1 AND 120),
 timezone text NOT NULL, locale text NOT NULL CHECK(locale IN ('en','hi')),
 state text NOT NULL DEFAULT 'active' CHECK(state IN ('active','suspended','closed')),
 version integer NOT NULL DEFAULT 1, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE campus (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id),
 name text NOT NULL CHECK(length(name) BETWEEN 1 AND 120), code text NOT NULL,
 address text NOT NULL DEFAULT '', state text NOT NULL DEFAULT 'active' CHECK(state IN ('active','closed')),
 version integer NOT NULL DEFAULT 1, UNIQUE(institution_id,code), UNIQUE(institution_id,id)
);
CREATE TABLE membership (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id),
 person_id text NOT NULL REFERENCES person(id), campus_id text,
 role text NOT NULL CHECK(role IN ('owner','coordinator','instructor','learner','guardian','reviewer')),
 state text NOT NULL DEFAULT 'active' CHECK(state IN ('active','suspended','archived')),
 version integer NOT NULL DEFAULT 1, created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),
 CHECK(role<>'owner' OR campus_id IS NULL),
 UNIQUE NULLS NOT DISTINCT(institution_id,person_id,campus_id,role)
);
CREATE INDEX membership_actor ON membership(person_id,institution_id) WHERE state='active';
CREATE TABLE session (
 id text PRIMARY KEY, person_id text NOT NULL REFERENCES person(id), token_hash text NOT NULL UNIQUE,
 csrf text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(), last_seen timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL, revoked_at timestamptz, device text NOT NULL, mfa_verified boolean NOT NULL
);
CREATE INDEX session_person ON session(person_id);
CREATE TABLE security_token (
 id text PRIMARY KEY, person_id text NOT NULL REFERENCES person(id), token_hash text NOT NULL UNIQUE,
 kind text NOT NULL CHECK(kind IN ('invitation','reset','verify')), payload jsonb NOT NULL DEFAULT '{}',
 expires_at timestamptz NOT NULL, consumed_at timestamptz, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE recovery_code (person_id text NOT NULL REFERENCES person(id), code_hash text NOT NULL, used_at timestamptz, PRIMARY KEY(person_id,code_hash));
CREATE TABLE throttle (key text PRIMARY KEY, window_at timestamptz NOT NULL, attempts integer NOT NULL);
CREATE TABLE course (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id), campus_id text NOT NULL,
 name text NOT NULL, academic_year text NOT NULL, state text NOT NULL DEFAULT 'active',
 UNIQUE(institution_id,id), FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id)
);
CREATE SEQUENCE learner_reference_seq;
CREATE TABLE enrollment (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id), person_id text NOT NULL REFERENCES person(id),
 campus_id text NOT NULL, course_id text NOT NULL,
 reference text NOT NULL UNIQUE DEFAULT ('L-' || lpad(nextval('learner_reference_seq')::text,8,'0')),
 state text NOT NULL CHECK(state IN ('pending','active','withdrawn')), start_date date NOT NULL,
 end_date date, version integer NOT NULL DEFAULT 1,
 FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),
 FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id),
 UNIQUE(institution_id,person_id,course_id)
);
CREATE TABLE family_link (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id),
 guardian_id text NOT NULL REFERENCES person(id), learner_id text NOT NULL REFERENCES person(id),
 relationship text NOT NULL, state text NOT NULL CHECK(state IN ('approved','removed')),
 approved_by text NOT NULL REFERENCES person(id), created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,guardian_id,learner_id), CHECK(guardian_id<>learner_id)
);
CREATE TABLE responsibility (
 id text PRIMARY KEY, institution_id text NOT NULL REFERENCES institution(id), staff_id text NOT NULL REFERENCES person(id),
 course_id text NOT NULL, subject text NOT NULL, starts_on date NOT NULL, ends_on date,
 FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id), CHECK(ends_on IS NULL OR ends_on>=starts_on),
 UNIQUE(institution_id,staff_id,course_id,subject)
);
CREATE TABLE audit_entry (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, institution_id text REFERENCES institution(id), actor_id text REFERENCES person(id),
 action text NOT NULL, resource_id text NOT NULL, reason text NOT NULL DEFAULT '', request_id text NOT NULL,
 occurred_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_institution_time ON audit_entry(institution_id,id DESC);
CREATE TABLE delivery_job (
 id text PRIMARY KEY, institution_id text REFERENCES institution(id), recipient_id text NOT NULL REFERENCES person(id),
 kind text NOT NULL, payload jsonb NOT NULL, state text NOT NULL DEFAULT 'queued' CHECK(state IN ('queued','running','accepted','failed','cancelled')),
 attempts integer NOT NULL DEFAULT 0, run_at timestamptz NOT NULL DEFAULT now(), lease_until timestamptz,
 lease_token text, last_error text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX delivery_ready ON delivery_job(run_at) WHERE state IN ('queued','running');

-- +goose Down
DROP TABLE delivery_job, audit_entry, responsibility, family_link, enrollment;
DROP SEQUENCE learner_reference_seq;
DROP TABLE course, throttle, recovery_code, security_token, session, membership, campus, institution, person;
