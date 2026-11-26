-- +goose Up
DELETE FROM tasks;
ALTER TABLE tasks
    ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX tasks_user_id_idx ON tasks (user_id);

-- +goose Down
DROP INDEX tasks_user_id_idx;
ALTER TABLE tasks DROP COLUMN user_id;