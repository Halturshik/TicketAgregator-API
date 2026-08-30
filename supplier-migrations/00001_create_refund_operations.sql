-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS supplier_simulator;

CREATE TABLE supplier_simulator.refund_operations (
	id UUID PRIMARY KEY,
	provider_code VARCHAR(40) NOT NULL,
	idempotency_key UUID NOT NULL,
	request_hash VARCHAR(64) NOT NULL,
	status VARCHAR(20) NOT NULL,
	failure_code VARCHAR(50),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT supplier_refund_operations_status_check
		CHECK (status IN ('success', 'rejected')),
	CONSTRAINT supplier_refund_operations_idempotency_unique
		UNIQUE (provider_code, idempotency_key)
);

CREATE TABLE supplier_simulator.refund_items (
	id BIGSERIAL PRIMARY KEY,
	operation_id UUID NOT NULL REFERENCES supplier_simulator.refund_operations(id) ON DELETE CASCADE,
	provider_code VARCHAR(40) NOT NULL,
	aggregator_ticket_id INT NOT NULL,
	ticket_number VARCHAR(20) NOT NULL,
	supplier_offer_id UUID NOT NULL,
	fare_type VARCHAR(30) NOT NULL,
	departure_at TIMESTAMPTZ NOT NULL,
	gross_amount INT NOT NULL CHECK (gross_amount > 0),
	refunded BOOLEAN NOT NULL,
	reason VARCHAR(50) NOT NULL,
	refund_percent INT NOT NULL CHECK (refund_percent BETWEEN 0 AND 100),
	refund_amount INT NOT NULL CHECK (refund_amount >= 0),
	refund_policy JSONB NOT NULL,
	CONSTRAINT supplier_refund_items_operation_ticket_unique
		UNIQUE (operation_id, aggregator_ticket_id)
);

CREATE UNIQUE INDEX supplier_refund_items_refunded_ticket_unique
	ON supplier_simulator.refund_items (provider_code, supplier_offer_id, ticket_number)
	WHERE refunded;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP SCHEMA supplier_simulator CASCADE;

-- +goose StatementEnd
