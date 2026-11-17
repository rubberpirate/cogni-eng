-- +goose Up
ALTER TABLE question_record DROP CONSTRAINT question_record_institution_id_fingerprint_key;
ALTER TABLE question_record ADD UNIQUE(institution_id,topic_id,fingerprint);
-- +goose Down
-- Cross-topic copies are intentional independent records. Refuse rollback if they would violate the older boundary.
ALTER TABLE question_record DROP CONSTRAINT question_record_institution_id_topic_id_fingerprint_key;
ALTER TABLE question_record ADD UNIQUE(institution_id,fingerprint);
