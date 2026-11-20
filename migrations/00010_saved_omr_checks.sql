-- +goose Up
CREATE TABLE omr_template (
 id text PRIMARY KEY,
 institution_id text NOT NULL REFERENCES institution(id),
 campus_id text NOT NULL,
 course_id text NOT NULL,
 name text NOT NULL CHECK(length(name) BETWEEN 1 AND 80),
 question_count integer NOT NULL CHECK(question_count BETWEEN 1 AND 200),
 definition jsonb NOT NULL,
 reference_asset_id text NOT NULL,
 created_by text NOT NULL REFERENCES person(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),
 FOREIGN KEY(institution_id,campus_id,course_id) REFERENCES course(institution_id,campus_id,id),
 FOREIGN KEY(institution_id,reference_asset_id) REFERENCES academic_asset(institution_id,id)
);
CREATE UNIQUE INDEX omr_template_name ON omr_template(institution_id,course_id,lower(name));
CREATE TABLE omr_check (
 id text PRIMARY KEY,
 institution_id text NOT NULL REFERENCES institution(id),
 paper_id text NOT NULL,
 paper_version integer NOT NULL,
 template_id text NOT NULL,
 source_asset_id text NOT NULL,
 sha256 text NOT NULL,
 job_id text NOT NULL,
 created_by text NOT NULL REFERENCES person(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),
 UNIQUE(institution_id,id,paper_id),
 UNIQUE(institution_id,paper_id,template_id,sha256),
 FOREIGN KEY(institution_id,paper_id) REFERENCES paper(institution_id,id),
 FOREIGN KEY(institution_id,template_id) REFERENCES omr_template(institution_id,id),
 FOREIGN KEY(institution_id,source_asset_id) REFERENCES academic_asset(institution_id,id),
 FOREIGN KEY(institution_id,job_id) REFERENCES academic_job(institution_id,id)
);
CREATE INDEX omr_check_paper ON omr_check(institution_id,paper_id,created_at DESC);
CREATE TABLE omr_checked_sheet (
 id text PRIMARY KEY,
 institution_id text NOT NULL REFERENCES institution(id),
 check_id text NOT NULL,
 paper_id text NOT NULL,
 page integer NOT NULL CHECK(page BETWEEN 1 AND 20),
 asset_id text NOT NULL,
 registration_number text NOT NULL DEFAULT '',
 student_name text NOT NULL DEFAULT '',
 reading jsonb NOT NULL,
 answers jsonb NOT NULL,
 result jsonb NOT NULL,
 uncertain_questions jsonb NOT NULL DEFAULT '[]',
 identity_uncertain boolean NOT NULL,
 state text NOT NULL CHECK(state IN ('needs_review','graded','excluded')),
 version integer NOT NULL DEFAULT 1,
 reviews jsonb NOT NULL DEFAULT '[]',
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),
 UNIQUE(institution_id,check_id,page),
 FOREIGN KEY(institution_id,check_id,paper_id) REFERENCES omr_check(institution_id,id,paper_id),
 FOREIGN KEY(institution_id,asset_id) REFERENCES academic_asset(institution_id,id)
);
CREATE INDEX omr_checked_sheet_paper ON omr_checked_sheet(institution_id,paper_id,registration_number);
CREATE UNIQUE INDEX omr_checked_sheet_registration ON omr_checked_sheet(institution_id,paper_id,lower(registration_number)) WHERE state='graded' AND registration_number<>'';
-- +goose Down
DROP TABLE omr_checked_sheet;
DROP TABLE omr_check;
DROP TABLE omr_template;
