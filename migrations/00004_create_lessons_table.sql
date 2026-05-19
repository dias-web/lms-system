-- +goose Up
-- +goose StatementBegin
CREATE TABLE lessons (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    content     TEXT         NOT NULL DEFAULT '',
    "order"     INTEGER      NOT NULL DEFAULT 0,
    chapter_id  BIGINT       NOT NULL REFERENCES chapters (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lessons_chapter_id ON lessons (chapter_id);
CREATE INDEX idx_lessons_chapter_order ON lessons (chapter_id, "order");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lessons;
-- +goose StatementEnd