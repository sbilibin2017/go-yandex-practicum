-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE IF NOT EXISTS content.metrics (
    id VARCHAR(255),
    type VARCHAR(255),
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, type)  
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS content.metrics;
DROP SCHEMA IF EXISTS content;
-- +goose StatementEnd
