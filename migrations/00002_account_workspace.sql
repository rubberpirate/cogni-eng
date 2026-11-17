-- +goose Up
ALTER TABLE person ADD COLUMN phone text NOT NULL DEFAULT '';
ALTER TABLE person ADD COLUMN emergency_contact text NOT NULL DEFAULT '';
ALTER TABLE institution ADD COLUMN support_email text NOT NULL DEFAULT '';
ALTER TABLE institution ADD COLUMN academic_year text NOT NULL DEFAULT '';
ALTER TABLE institution ADD COLUMN require_staff_mfa boolean NOT NULL DEFAULT false;
CREATE TABLE identity_change (
 id text PRIMARY KEY, person_id text NOT NULL REFERENCES person(id), new_email text NOT NULL,
 token_hash text NOT NULL UNIQUE, expires_at timestamptz NOT NULL, consumed_at timestamptz
);
CREATE TABLE placement_event (
 id text PRIMARY KEY, enrollment_id text NOT NULL REFERENCES enrollment(id), institution_id text NOT NULL REFERENCES institution(id),
 previous_campus_id text NOT NULL, previous_course_id text NOT NULL, previous_state text NOT NULL,
 next_campus_id text NOT NULL, next_course_id text NOT NULL, next_state text NOT NULL,
 effective_on date NOT NULL, actor_id text NOT NULL REFERENCES person(id), reason text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE delivery_job DROP CONSTRAINT delivery_job_state_check;
ALTER TABLE delivery_job ADD CONSTRAINT delivery_job_state_check CHECK(state IN ('queued','running','accepted','failed','cancelled'));
-- +goose Down
DROP TABLE placement_event,identity_change;
ALTER TABLE institution DROP COLUMN require_staff_mfa,DROP COLUMN academic_year,DROP COLUMN support_email;
ALTER TABLE person DROP COLUMN emergency_contact,DROP COLUMN phone;
