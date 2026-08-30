-- +goose Up
-- +goose StatementBegin

ALTER TABLE users
	ADD COLUMN bonus_debt INT NOT NULL DEFAULT 0 CHECK (bonus_debt >= 0);

ALTER TABLE tickets
	ADD COLUMN supplier_code VARCHAR(40),
	ADD COLUMN supplier_offer_id VARCHAR(64),
	ADD COLUMN fare_type VARCHAR(30),
	ADD COLUMN refund_policy_version INT,
	ADD COLUMN refund_policy_snapshot JSONB,
	ADD COLUMN bonus_spent INT,
	ADD COLUMN bonus_earned INT,
	ADD COLUMN payable_amount INT;

UPDATE tickets
SET supplier_code = 'legacy',
	supplier_offer_id = ticket_number,
	fare_type = 'non_refundable',
	refund_policy_version = 1,
	refund_policy_snapshot = '{"code":"non_refundable","refundable":false}'::JSONB,
	bonus_spent = 0,
	bonus_earned = 0,
	payable_amount = price;

ALTER TABLE tickets
	ALTER COLUMN supplier_code SET NOT NULL,
	ALTER COLUMN supplier_offer_id SET NOT NULL,
	ALTER COLUMN fare_type SET NOT NULL,
	ALTER COLUMN refund_policy_version SET NOT NULL,
	ALTER COLUMN refund_policy_snapshot SET NOT NULL,
	ALTER COLUMN bonus_spent SET NOT NULL,
	ALTER COLUMN bonus_earned SET NOT NULL,
	ALTER COLUMN payable_amount SET NOT NULL,
	ADD CONSTRAINT tickets_fare_type_check CHECK (fare_type IN ('non_refundable', 'standard', 'flexible')),
	ADD CONSTRAINT tickets_bonus_spent_check CHECK (bonus_spent >= 0),
	ADD CONSTRAINT tickets_bonus_earned_check CHECK (bonus_earned >= 0),
	ADD CONSTRAINT tickets_payable_amount_check CHECK (payable_amount >= 0),
	ADD CONSTRAINT tickets_financials_check CHECK (bonus_spent + payable_amount = price);

ALTER TABLE tickets DROP CONSTRAINT tickets_status_check;
ALTER TABLE tickets
	ADD CONSTRAINT tickets_status_check CHECK (status IN ('booked', 'paid', 'cancelled', 'expired', 'refund_pending', 'refunded'));

ALTER TABLE orders DROP CONSTRAINT orders_status_check;
ALTER TABLE orders
	ADD CONSTRAINT orders_status_check CHECK (status IN ('created', 'paid', 'cancelled', 'expired', 'partially_refunded', 'refunded'));

ALTER TABLE bonus_transactions DROP CONSTRAINT bonus_transactions_type_check;
ALTER TABLE bonus_transactions
	ADD CONSTRAINT bonus_transactions_type_check CHECK (
		type IN ('earn', 'spend', 'rollback', 'refund_restore', 'refund_revoke', 'debt_create', 'debt_repay')
	);

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
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ,
	CONSTRAINT refunds_status_check CHECK (status IN ('processing', 'success', 'failed'))
);

ALTER TABLE bonus_transactions
	ADD COLUMN refund_id INT REFERENCES refunds(id) ON DELETE SET NULL;

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
	cash_refunded INT NOT NULL CHECK (cash_refunded >= 0),
	bonus_restored INT NOT NULL CHECK (bonus_restored >= 0),
	bonus_revoked INT NOT NULL CHECK (bonus_revoked >= 0),
	CONSTRAINT refund_items_operation_ticket_unique UNIQUE (refund_id, ticket_id)
);

CREATE INDEX refunds_order_created_idx ON refunds (order_id, created_at DESC);
CREATE UNIQUE INDEX refunds_supplier_operation_unique
	ON refunds (supplier_code, supplier_refund_id)
	WHERE supplier_refund_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE refund_items;
ALTER TABLE bonus_transactions DROP COLUMN refund_id;
DROP TABLE refunds;

ALTER TABLE bonus_transactions DROP CONSTRAINT bonus_transactions_type_check;
ALTER TABLE bonus_transactions
	ADD CONSTRAINT bonus_transactions_type_check CHECK (type IN ('earn', 'spend', 'rollback'));

ALTER TABLE orders DROP CONSTRAINT orders_status_check;
ALTER TABLE orders
	ADD CONSTRAINT orders_status_check CHECK (status IN ('created', 'paid', 'cancelled', 'expired'));

ALTER TABLE tickets DROP CONSTRAINT tickets_status_check;
ALTER TABLE tickets
	ADD CONSTRAINT tickets_status_check CHECK (status IN ('booked', 'paid', 'cancelled', 'expired'));

ALTER TABLE tickets
	DROP CONSTRAINT tickets_financials_check,
	DROP CONSTRAINT tickets_payable_amount_check,
	DROP CONSTRAINT tickets_bonus_earned_check,
	DROP CONSTRAINT tickets_bonus_spent_check,
	DROP CONSTRAINT tickets_fare_type_check,
	DROP COLUMN payable_amount,
	DROP COLUMN bonus_earned,
	DROP COLUMN bonus_spent,
	DROP COLUMN refund_policy_snapshot,
	DROP COLUMN refund_policy_version,
	DROP COLUMN fare_type,
	DROP COLUMN supplier_offer_id,
	DROP COLUMN supplier_code;

ALTER TABLE users DROP COLUMN bonus_debt;

-- +goose StatementEnd
