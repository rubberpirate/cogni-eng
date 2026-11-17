-- +goose Up
CREATE SEQUENCE academic_reference_seq;
CREATE TABLE topic (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,course_id text NOT NULL,
 subject text NOT NULL,name text NOT NULL,outcome text NOT NULL,language text NOT NULL DEFAULT 'en',
 UNIQUE(institution_id,id),UNIQUE(course_id,subject,name),
 FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id)
);
CREATE TABLE question_record (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),topic_id text NOT NULL,
 reference text NOT NULL UNIQUE DEFAULT ('Q-'||lpad(nextval('academic_reference_seq')::text,8,'0')),fingerprint text NOT NULL,
 latest_revision text,version integer NOT NULL DEFAULT 1,created_by text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),UNIQUE(institution_id,fingerprint),FOREIGN KEY(institution_id,topic_id) REFERENCES topic(institution_id,id)
);
CREATE TABLE question_revision (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),record_id text NOT NULL,
 content jsonb NOT NULL,state text NOT NULL CHECK(state IN ('draft','review','approved','retired')),
 origin text NOT NULL CHECK(origin IN ('authored','imported','generated','development_fixture')),
 created_by text NOT NULL REFERENCES person(id),reviewed_by text REFERENCES person(id),review_note text NOT NULL DEFAULT '',created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,record_id) REFERENCES question_record(institution_id,id)
);
ALTER TABLE question_record ADD FOREIGN KEY(institution_id,latest_revision) REFERENCES question_revision(institution_id,id);
CREATE TABLE academic_asset (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,course_id text NOT NULL,
 owner_id text NOT NULL REFERENCES person(id),purpose text NOT NULL CHECK(purpose IN ('question_image','document','scan','scan_page')),
 object_key text NOT NULL UNIQUE,filename text NOT NULL,media_type text NOT NULL,bytes bigint NOT NULL CHECK(bytes BETWEEN 1 AND 52428800),sha256 text NOT NULL,
 state text NOT NULL DEFAULT 'staged' CHECK(state IN ('staged','ready','failed','retired')),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id)
);
CREATE TABLE question_import (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),topic_id text NOT NULL,asset_id text,
 fingerprint text NOT NULL,state text NOT NULL CHECK(state IN ('preview','committed','extracting','failed','cancelled')),rows jsonb NOT NULL DEFAULT '[]',raw_text text NOT NULL DEFAULT '',
 version integer NOT NULL DEFAULT 1,created_by text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),UNIQUE(institution_id,topic_id,fingerprint),FOREIGN KEY(institution_id,topic_id) REFERENCES topic(institution_id,id),FOREIGN KEY(institution_id,asset_id) REFERENCES academic_asset(institution_id,id)
);
CREATE TABLE paper (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,course_id text NOT NULL,
 reference text NOT NULL UNIQUE DEFAULT ('P-'||lpad(nextval('academic_reference_seq')::text,8,'0')),
 spec jsonb NOT NULL,state text NOT NULL DEFAULT 'draft' CHECK(state IN ('draft','published','archived')),
 version integer NOT NULL DEFAULT 1,created_by text NOT NULL REFERENCES person(id),published_by text REFERENCES person(id),published_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id)
);
CREATE TABLE assignment (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),paper_id text NOT NULL,learner_id text NOT NULL REFERENCES person(id),
 opens_at timestamptz NOT NULL,closes_at timestamptz NOT NULL,extra_minutes integer NOT NULL DEFAULT 0 CHECK(extra_minutes BETWEEN 0 AND 480),
 assigned_by text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),UNIQUE(paper_id,learner_id),CHECK(closes_at>opens_at),FOREIGN KEY(institution_id,paper_id) REFERENCES paper(institution_id,id)
);
CREATE TABLE sitting (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),assignment_id text NOT NULL,learner_id text NOT NULL REFERENCES person(id),
 channel text NOT NULL CHECK(channel IN ('online','paper')),state text NOT NULL CHECK(state IN ('active','awaiting_review','marked','released')),
 started_at timestamptz NOT NULL,deadline timestamptz NOT NULL,finished_at timestamptz,received_at timestamptz NOT NULL DEFAULT now(),finish_reason text NOT NULL DEFAULT '',
 writer_session text REFERENCES session(id),answers jsonb NOT NULL DEFAULT '{}',manual_marks jsonb NOT NULL DEFAULT '{}',answer_version integer NOT NULL DEFAULT 0,
 result jsonb,marking_revision integer NOT NULL DEFAULT 0,release_at timestamptz,version integer NOT NULL DEFAULT 1,
 UNIQUE(institution_id,id),UNIQUE(assignment_id),FOREIGN KEY(institution_id,assignment_id) REFERENCES assignment(institution_id,id)
);
CREATE TABLE response_operation (sitting_id text NOT NULL REFERENCES sitting(id),operation_id text NOT NULL,version integer NOT NULL,PRIMARY KEY(sitting_id,operation_id));
CREATE TABLE result_revision (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),sitting_id text NOT NULL,revision integer NOT NULL,result jsonb NOT NULL,
 reason text NOT NULL,actor_id text REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(sitting_id,revision),FOREIGN KEY(institution_id,sitting_id) REFERENCES sitting(institution_id,id)
);
CREATE TABLE academic_job (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),campus_id text NOT NULL,course_id text NOT NULL,actor_id text NOT NULL REFERENCES person(id),
 kind text NOT NULL CHECK(kind IN ('generate_questions','recognize_scan','extract_document','tutor')),
 resource_id text NOT NULL,input jsonb NOT NULL,state text NOT NULL DEFAULT 'queued' CHECK(state IN ('queued','running','completed','failed','cancelled','configuration_required')),
 progress integer NOT NULL DEFAULT 0 CHECK(progress BETWEEN 0 AND 100),error text NOT NULL DEFAULT '',attempts integer NOT NULL DEFAULT 0,
 output jsonb NOT NULL DEFAULT '{}',river_id bigint,provider text NOT NULL DEFAULT '',model text NOT NULL DEFAULT '',input_tokens bigint,output_tokens bigint,
 created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,campus_id) REFERENCES campus(institution_id,id),FOREIGN KEY(institution_id,course_id) REFERENCES course(institution_id,id)
);
CREATE TABLE sheet_layout (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),paper_id text NOT NULL,name text NOT NULL,spec jsonb NOT NULL,fingerprint text NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,paper_id) REFERENCES paper(institution_id,id)
);
CREATE TABLE scan_batch (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),paper_id text NOT NULL,layout_id text NOT NULL,sitting_at timestamptz NOT NULL,created_by text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,paper_id) REFERENCES paper(institution_id,id),FOREIGN KEY(institution_id,layout_id) REFERENCES sheet_layout(institution_id,id)
);
CREATE TABLE scan_source (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),batch_id text NOT NULL,asset_id text NOT NULL,job_id text NOT NULL,sha256 text NOT NULL,
 UNIQUE(institution_id,id),UNIQUE(batch_id,sha256),FOREIGN KEY(institution_id,batch_id) REFERENCES scan_batch(institution_id,id),FOREIGN KEY(institution_id,asset_id) REFERENCES academic_asset(institution_id,id),FOREIGN KEY(institution_id,job_id) REFERENCES academic_job(institution_id,id)
);
CREATE TABLE scan_sheet (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),batch_id text NOT NULL,source_id text NOT NULL,asset_id text NOT NULL,assignment_id text,
 page integer NOT NULL,sha256 text NOT NULL,machine jsonb NOT NULL,corrections jsonb NOT NULL DEFAULT '{}',identity_confirmed boolean NOT NULL DEFAULT false,
 state text NOT NULL CHECK(state IN ('review','ready','committed','rejected')),version integer NOT NULL DEFAULT 1,committed_sitting text REFERENCES sitting(id),
 UNIQUE(institution_id,id),UNIQUE(batch_id,sha256),UNIQUE(source_id,page),FOREIGN KEY(institution_id,batch_id) REFERENCES scan_batch(institution_id,id),FOREIGN KEY(institution_id,asset_id) REFERENCES academic_asset(institution_id,id),FOREIGN KEY(institution_id,assignment_id) REFERENCES assignment(institution_id,id)
);
CREATE TABLE scan_correction (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),sheet_id text NOT NULL,question_id text NOT NULL,previous_value jsonb NOT NULL,new_value jsonb NOT NULL,
 reason text NOT NULL,actor_id text NOT NULL REFERENCES person(id),created_at timestamptz NOT NULL DEFAULT now(),FOREIGN KEY(institution_id,sheet_id) REFERENCES scan_sheet(institution_id,id)
);
CREATE TABLE answer_exposure (institution_id text NOT NULL REFERENCES institution(id),learner_id text NOT NULL REFERENCES person(id),question_id text NOT NULL,first_seen timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(institution_id,learner_id,question_id));
CREATE TABLE tutor_thread (
 id text PRIMARY KEY,institution_id text NOT NULL REFERENCES institution(id),learner_id text NOT NULL REFERENCES person(id),sitting_id text NOT NULL,question_id text NOT NULL,messages jsonb NOT NULL DEFAULT '[]',version integer NOT NULL DEFAULT 1,
 UNIQUE(institution_id,id),FOREIGN KEY(institution_id,sitting_id) REFERENCES sitting(institution_id,id)
);
CREATE INDEX sitting_active_deadline ON sitting(deadline) WHERE state='active';
CREATE INDEX academic_job_tenant_time ON academic_job(institution_id,created_at DESC);
-- +goose Down
DROP TABLE tutor_thread,answer_exposure,scan_correction,scan_sheet,scan_source,scan_batch,sheet_layout,academic_job,result_revision,response_operation,sitting,assignment,paper,question_import,academic_asset;
ALTER TABLE question_record DROP CONSTRAINT question_record_institution_id_latest_revision_fkey;
DROP TABLE question_revision,question_record,topic;
DROP SEQUENCE academic_reference_seq;
