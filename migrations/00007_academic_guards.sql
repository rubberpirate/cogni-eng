-- +goose Up
ALTER TABLE course ADD UNIQUE(institution_id,campus_id,id);
ALTER TABLE topic ADD FOREIGN KEY(institution_id,campus_id,course_id) REFERENCES course(institution_id,campus_id,id);
ALTER TABLE academic_asset ADD FOREIGN KEY(institution_id,campus_id,course_id) REFERENCES course(institution_id,campus_id,id);
ALTER TABLE paper ADD FOREIGN KEY(institution_id,campus_id,course_id) REFERENCES course(institution_id,campus_id,id);
-- +goose StatementBegin
CREATE FUNCTION preserve_academic_content() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='question_revision' THEN
  IF NEW.content IS DISTINCT FROM OLD.content OR NEW.record_id IS DISTINCT FROM OLD.record_id OR NEW.institution_id IS DISTINCT FROM OLD.institution_id THEN
   RAISE EXCEPTION 'question revision content is immutable; create a new revision' USING ERRCODE='23514';
  END IF;
 ELSE
  IF OLD.state<>'draft' AND (NEW.spec IS DISTINCT FROM OLD.spec OR NEW.institution_id IS DISTINCT FROM OLD.institution_id OR NEW.campus_id IS DISTINCT FROM OLD.campus_id OR NEW.course_id IS DISTINCT FROM OLD.course_id) THEN
   RAISE EXCEPTION 'published paper content is immutable' USING ERRCODE='23514';
  END IF;
 END IF;
 RETURN NEW;
END;
$$;
-- +goose StatementEnd
CREATE TRIGGER preserve_question_revision BEFORE UPDATE ON question_revision FOR EACH ROW EXECUTE FUNCTION preserve_academic_content();
CREATE TRIGGER preserve_published_paper BEFORE UPDATE ON paper FOR EACH ROW EXECUTE FUNCTION preserve_academic_content();
-- +goose Down
DROP TRIGGER preserve_question_revision ON question_revision;
DROP TRIGGER preserve_published_paper ON paper;
DROP FUNCTION preserve_academic_content();
ALTER TABLE paper DROP CONSTRAINT paper_institution_id_campus_id_course_id_fkey;
ALTER TABLE academic_asset DROP CONSTRAINT academic_asset_institution_id_campus_id_course_id_fkey;
ALTER TABLE topic DROP CONSTRAINT topic_institution_id_campus_id_course_id_fkey;
ALTER TABLE course DROP CONSTRAINT course_institution_id_campus_id_id_key;
