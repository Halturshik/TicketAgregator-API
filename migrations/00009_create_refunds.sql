-- +goose Up
-- +goose StatementBegin

CREATE TABLE refunds (
	id SERIAL PRIMARY KEY,
	order_id INT NOT NULL REFERENCES orders(id),
	payment_id INT NOT NULL REFERENCES payments(id),
	status VARCHAR(20) NOT NULL,
	idempotency_key UUID NOT NULL UNIQUE,
	request_hash VARCHAR(64) NOT NULL,
	supplier_code VARCHAR(40) NOT NULL,
	supplier_refund_id UUID,
	cash_amount INT NOT NULL CHECK (cash_amount >= 0),
	bonus_restored INT NOT NULL CHECK (bonus_restored >= 0),
	bonus_revoked INT NOT NULL CHECK (bonus_revoked >= 0),
	bonus_debt_created INT NOT NULL DEFAULT 0 CHECK (bonus_debt_created >= 0),
	bonus_balance_after INT CHECK (bonus_balance_after >= 0),
	bonus_debt_after INT CHECK (bonus_debt_after >= 0),
	attempt_count INT NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
	failure_code VARCHAR(50),
	last_error VARCHAR(500),
	next_retry_at TIMESTAMPTZ NOT NULL,
	reconciliation_deadline TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ,

	CONSTRAINT refunds_status_check CHECK (
		status IN ('processing', 'success', 'failed', 'requires_review')
	),
	CONSTRAINT refunds_reconciliation_time_check CHECK (
		reconciliation_deadline >= next_retry_at
	)
);

CREATE TABLE refund_items (
	id SERIAL PRIMARY KEY,
	refund_id INT NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
	ticket_id INT NOT NULL REFERENCES tickets(id),
	supplier_code VARCHAR(40) NOT NULL,
	ticket_number VARCHAR(20) NOT NULL,
	supplier_offer_id UUID NOT NULL,
	fare_type VARCHAR(30) NOT NULL,
	departure_at TIMESTAMPTZ NOT NULL,
	reason VARCHAR(50) NOT NULL,
	refund_percent INT NOT NULL CHECK (refund_percent BETWEEN 0 AND 100),
	gross_amount INT NOT NULL CHECK (gross_amount > 0),
	supplier_refund_amount INT NOT NULL CHECK (supplier_refund_amount >= 0),

	CONSTRAINT refund_items_operation_ticket_unique UNIQUE (refund_id, ticket_id)
);

CREATE TABLE bonus_transactions (
	id SERIAL PRIMARY KEY,
	user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	order_id INT REFERENCES orders(id) ON DELETE SET NULL,
	refund_id INT REFERENCES refunds(id) ON DELETE SET NULL,
	type VARCHAR(20) NOT NULL,
	amount INT NOT NULL CHECK (amount > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	CONSTRAINT bonus_transactions_type_check CHECK (
		type IN ('earn', 'spend', 'rollback', 'refund_restore', 'refund_revoke', 'debt_create', 'debt_repay')
	)
);

CREATE INDEX refunds_order_created_idx ON refunds (order_id, created_at DESC);

CREATE UNIQUE INDEX refunds_supplier_operation_unique
	ON refunds (supplier_code, supplier_refund_id)
	WHERE supplier_refund_id IS NOT NULL;

CREATE INDEX refunds_reconciliation_due_idx
	ON refunds (next_retry_at, id)
	WHERE status = 'processing';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE bonus_transactions;
DROP TABLE refund_items;
DROP TABLE refunds;

-- +goose StatementEnd
