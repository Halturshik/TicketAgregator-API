-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_internal_passport
) VALUES ('any', false, 14, 150, true, true);

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_birth_certificate
) VALUES ('any', false, 0, 13, true, true);

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_foreign_passport
) VALUES ('any', false, 0, 150, false, true);

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_international_passport
) VALUES ('any', true, 14, 150, true, true);

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_international_passport
) VALUES ('any', true, 0, 13, true, true);

INSERT INTO document_rules (
    transport_type, is_international, min_age, max_age, is_russian,
    allow_foreign_passport
) VALUES ('any', true, 0, 150, false, true);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DELETE FROM document_rules;
