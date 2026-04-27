-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,

    owner_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    passenger_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    passenger_relative_id INT REFERENCES user_relatives(id) ON DELETE SET NULL,

    guest_email VARCHAR(255),

    transport_type VARCHAR(20) NOT NULL,
    is_international BOOLEAN NOT NULL,

    departure_city VARCHAR(100) NOT NULL,
    arrival_city VARCHAR(100) NOT NULL,
    departure_time TIMESTAMP NOT NULL,

    document_number VARCHAR(50) NOT NULL,

    price INT NOT NULL,
    status VARCHAR(20) DEFAULT 'booked'
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE tickets;
