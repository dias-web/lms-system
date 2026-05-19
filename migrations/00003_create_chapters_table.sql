-- +goose Up
-- +goose StatementBegin
CREATE TABLE chapters (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    "order"     INTEGER      NOT NULL DEFAULT 0,
    course_id   BIGINT       NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chapters_course_id ON chapters (course_id);
CREATE INDEX idx_chapters_course_order ON chapters (course_id, "order");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chapters;
-- +goose StatementEnd