-- +goose Up
CREATE UNIQUE INDEX tutor_one_conversation ON tutor_thread(institution_id,learner_id,sitting_id,question_id);
-- +goose Down
DROP INDEX tutor_one_conversation;
