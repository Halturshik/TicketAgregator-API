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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ
);

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    owner_user_id INT REFERENCES users(id) ON DELETE CASCADE,
    passenger_id INT REFERENCES saved_passengers(id) ON DELETE CASCADE,

    type VARCHAR(40) NOT NULL,
    number VARCHAR(80) NOT NULL,
	verification_status VARCHAR(20) NOT NULL,
	document_fingerprint VARCHAR(64) NOT NULL,
    expires_at DATE,
	last_checked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT documents_one_owner CHECK (
        (owner_user_id IS NOT NULL AND passenger_id IS NULL)
        OR (owner_user_id IS NULL AND passenger_id IS NOT NULL)
    ),
    CONSTRAINT documents_status_check CHECK (
		verification_status IN ('verified', 'rejected')
	)
);

CREATE UNIQUE INDEX documents_user_fingerprint_unique
	ON documents (owner_user_id, document_fingerprint)
	WHERE owner_user_id IS NOT NULL;

CREATE UNIQUE INDEX documents_passenger_fingerprint_unique
	ON documents (passenger_id, document_fingerprint)
	WHERE passenger_id IS NOT NULL;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE documents;
DROP TABLE saved_passengers;
