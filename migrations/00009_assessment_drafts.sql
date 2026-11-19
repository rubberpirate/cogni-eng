-- +goose Up
CREATE TABLE assessment_draft (
 id text PRIMARY KEY,
 institution_id text NOT NULL REFERENCES institution(id),
 campus_id text NOT NULL,
 course_id text NOT NULL,
 spec jsonb NOT NULL,
 version integer NOT NULL DEFAULT 1,
 state text NOT NULL DEFAULT 'draft' CHECK (state IN ('draft','published')),
 paper_id text REFERENCES paper(id),
 created_by text NOT NULL REFERENCES person(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(institution_id,campus_id,course_id) REFERENCES course(institution_id,campus_id,id)
);
CREATE INDEX assessment_draft_scope ON assessment_draft(institution_id,campus_id,course_id,updated_at DESC);
-- +goose Down
DROP TABLE assessment_draft;
