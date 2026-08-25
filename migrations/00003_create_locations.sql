-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE countries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_russia BOOLEAN NOT NULL,
    neighbor_country_ids INT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE cities (
    id SERIAL PRIMARY KEY,
    country_id INT REFERENCES countries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    is_air_hub BOOLEAN NOT NULL DEFAULT FALSE,
    latitude NUMERIC(9,6),
    longitude NUMERIC(9,6)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE cities;
DROP TABLE countries;
