-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    guest_email VARCHAR(255),
    guest_payment_token UUID UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    total_price INT NOT NULL CHECK (total_price >= 0),
    bonus_spent INT NOT NULL DEFAULT 0 CHECK (bonus_spent >= 0),
    bonus_earned INT NOT NULL DEFAULT 0 CHECK (bonus_earned >= 0),
    payable_amount INT NOT NULL CHECK (payable_amount >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMP,

    CONSTRAINT orders_status_check CHECK (status IN ('created', 'paid', 'cancelled')),
    CONSTRAINT orders_owner_check CHECK (user_id IS NOT NULL OR guest_email IS NOT NULL)
);

CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	ticket_number VARCHAR(12) NOT NULL UNIQUE,

    saved_passenger_id INT REFERENCES saved_passengers(id) ON DELETE SET NULL,
    passenger_snapshot JSONB NOT NULL,
    document_snapshot JSONB NOT NULL,
    transport_type VARCHAR(20) NOT NULL,
    is_international BOOLEAN NOT NULL,

    price INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'booked',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT tickets_transport_check CHECK (transport_type IN ('avia', 'rail', 'bus')),
    CONSTRAINT tickets_status_check CHECK (status IN ('booked', 'paid', 'cancelled'))
);

CREATE TABLE ticket_segments (
    id SERIAL PRIMARY KEY,
    ticket_id INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    segment_order INT NOT NULL,
    from_city_id INT,
    to_city_id INT,
    from_city VARCHAR(100) NOT NULL,
    to_city VARCHAR(100) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
	carrier_id INT REFERENCES carriers(id) ON DELETE SET NULL,
    carrier VARCHAR(100) NOT NULL,
    carrier_code VARCHAR(3) NOT NULL,
    route_number VARCHAR(12) NOT NULL
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    provider VARCHAR(40) NOT NULL DEFAULT 'mock',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'success', 'failed'))
);

CREATE TABLE bonus_transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id INT REFERENCES orders(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    amount INT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT bonus_transactions_type_check CHECK (type IN ('earn', 'spend', 'rollback'))
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE bonus_transactions;
DROP TABLE payments;
DROP TABLE ticket_segments;
DROP TABLE tickets;
DROP TABLE orders;
