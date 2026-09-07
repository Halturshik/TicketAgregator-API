-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
	order_number VARCHAR(9) NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    guest_email VARCHAR(255),
    guest_payment_token UUID UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    total_price INT NOT NULL CHECK (total_price >= 0),
	current_total_price INT NOT NULL CHECK (current_total_price >= 0 AND current_total_price <= total_price),
    bonus_spent INT NOT NULL DEFAULT 0 CHECK (bonus_spent >= 0),
    bonus_earned INT NOT NULL DEFAULT 0 CHECK (bonus_earned >= 0),
    payable_amount INT NOT NULL CHECK (payable_amount >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	paid_at TIMESTAMPTZ,
	expires_at TIMESTAMPTZ NOT NULL,

	CONSTRAINT orders_order_number_unique UNIQUE (order_number),
	CONSTRAINT orders_order_number_format_check CHECK (order_number ~ '^[A-Z]{3}-[0-9]{5}$'),
	CONSTRAINT orders_bonus_limit_check CHECK (bonus_spent <= current_total_price / 2),
	CONSTRAINT orders_status_check CHECK (
		status IN ('created', 'paid', 'expired', 'partially_refunded', 'refunded')
	),
	CONSTRAINT orders_owner_check CHECK (
		(user_id IS NOT NULL AND guest_email IS NULL AND guest_payment_token IS NULL)
		OR (user_id IS NULL AND guest_email IS NOT NULL AND guest_payment_token IS NOT NULL)
	)
);

CREATE TABLE order_passengers (
	id SERIAL PRIMARY KEY,
	order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	passenger_source VARCHAR(20) NOT NULL,
	saved_passenger_id INT REFERENCES saved_passengers(id) ON DELETE SET NULL,
	saved_document_id INT REFERENCES documents(id) ON DELETE SET NULL,
	save_changes BOOLEAN NOT NULL DEFAULT FALSE,
	passenger_snapshot JSONB NOT NULL,
	document_snapshot JSONB NOT NULL,
	document_fingerprint VARCHAR(64) NOT NULL,

	CONSTRAINT order_passengers_source_check CHECK (passenger_source IN ('self', 'saved', 'new')),
	CONSTRAINT order_passengers_order_unique UNIQUE (id, order_id)
);

CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	ticket_number VARCHAR(12) NOT NULL UNIQUE,
	order_passenger_id INT NOT NULL,
	supplier_code VARCHAR(40) NOT NULL,
	supplier_offer_id VARCHAR(64) NOT NULL,
	fare_type VARCHAR(30) NOT NULL,
	refund_policy_version INT NOT NULL,
	refund_policy_snapshot JSONB NOT NULL,
    transport_type VARCHAR(20) NOT NULL,
    is_international BOOLEAN NOT NULL,

	price INT NOT NULL CHECK (price > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'booked',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tickets_transport_check CHECK (transport_type IN ('avia', 'rail', 'bus')),
	CONSTRAINT tickets_fare_type_check CHECK (fare_type IN ('non_refundable', 'standard', 'flexible')),
	CONSTRAINT tickets_status_check CHECK (
		status IN ('booked', 'paid', 'expired', 'refund_pending', 'refunded')
	),
	CONSTRAINT tickets_order_passenger_fk FOREIGN KEY (order_passenger_id, order_id)
		REFERENCES order_passengers(id, order_id) ON DELETE CASCADE
);

CREATE TABLE scheduled_trips (
	id SERIAL PRIMARY KEY,
	trip_key VARCHAR(64) NOT NULL UNIQUE,
	transport_type VARCHAR(20) NOT NULL,
	from_city_id INT REFERENCES cities(id) ON DELETE SET NULL,
	to_city_id INT REFERENCES cities(id) ON DELETE SET NULL,
	from_city VARCHAR(100) NOT NULL,
	to_city VARCHAR(100) NOT NULL,
	departure_time TIMESTAMPTZ NOT NULL,
	arrival_time TIMESTAMPTZ NOT NULL,
	carrier_id INT REFERENCES carriers(id) ON DELETE SET NULL,
	carrier VARCHAR(100) NOT NULL,
	carrier_code VARCHAR(3) NOT NULL,
	route_number VARCHAR(12) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	CONSTRAINT scheduled_trips_transport_check CHECK (transport_type IN ('avia', 'rail', 'bus')),
	CONSTRAINT scheduled_trips_time_check CHECK (arrival_time > departure_time)
);

CREATE TABLE ticket_segments (
	id SERIAL PRIMARY KEY,
	ticket_id INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
	scheduled_trip_id INT NOT NULL REFERENCES scheduled_trips(id),
	segment_order INT NOT NULL,

	CONSTRAINT ticket_segments_order_unique UNIQUE (ticket_id, segment_order)
);

CREATE INDEX orders_user_created_idx ON orders (user_id, created_at DESC);
CREATE INDEX tickets_order_idx ON tickets (order_id);
CREATE INDEX ticket_segments_trip_idx ON ticket_segments (scheduled_trip_id);
CREATE INDEX scheduled_trips_route_departure_idx ON scheduled_trips (route_number, departure_time);
CREATE INDEX scheduled_trips_created_idx ON scheduled_trips (created_at, id);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount >= 0),
	status VARCHAR(20) NOT NULL,
    provider VARCHAR(40) NOT NULL DEFAULT 'mock',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	CONSTRAINT payments_status_check CHECK (status IN ('success', 'failed'))
);

CREATE UNIQUE INDEX payments_order_success_unique
	ON payments (order_id)
	WHERE status = 'success';

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE payments;
DROP TABLE ticket_segments;
DROP TABLE tickets;
DROP TABLE order_passengers;
DROP TABLE orders;
DROP TABLE scheduled_trips;
