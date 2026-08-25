-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE document_rules (
    id SERIAL PRIMARY KEY,

    transport_type VARCHAR(20) NOT NULL,
    is_international BOOLEAN NOT NULL,

    min_age INT NOT NULL,
    max_age INT NOT NULL,

    is_russian BOOLEAN NOT NULL,

    allow_internal_passport BOOLEAN DEFAULT FALSE,
    allow_international_passport BOOLEAN DEFAULT FALSE,
    allow_birth_certificate BOOLEAN DEFAULT FALSE,
    allow_foreign_passport BOOLEAN DEFAULT FALSE
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE document_rules;