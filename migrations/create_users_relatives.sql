-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE saved_passengers (
    id SERIAL PRIMARY KEY,
    owner_user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    birth_date DATE NOT NULL,
    is_russian BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    owner_user_id INT REFERENCES users(id) ON DELETE CASCADE,
    passenger_id INT REFERENCES saved_passengers(id) ON DELETE CASCADE,

    type VARCHAR(40) NOT NULL,
    number VARCHAR(80) NOT NULL,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT documents_one_owner CHECK (
        (owner_user_id IS NOT NULL AND passenger_id IS NULL)
        OR (owner_user_id IS NULL AND passenger_id IS NOT NULL)
    ),
    CONSTRAINT documents_status_check CHECK (
        verification_status IN ('pending', 'verified', 'rejected')
    )
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE documents;
DROP TABLE saved_passengers;
