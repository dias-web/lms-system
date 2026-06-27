-- +goose Up
-- +goose StatementBegin
CREATE TABLE attachments (
    id           BIGSERIAL    PRIMARY KEY,
    lesson_id    BIGINT       NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    file_name    VARCHAR(255) NOT NULL,
    object_key   VARCHAR(512) NOT NULL,
    content_type VARCHAR(255) NOT NULL DEFAULT '',
    size         BIGINT       NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_attachments_object_key ON attachments (object_key);
CREATE INDEX idx_attachments_lesson_id ON attachments (lesson_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS attachments;
-- +goose StatementEnd
