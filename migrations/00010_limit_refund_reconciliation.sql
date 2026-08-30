-- +goose Up
-- +goose StatementBegin

ALTER TABLE refunds
	ADD COLUMN next_retry_at TIMESTAMPTZ,
	ADD COLUMN reconciliation_deadline TIMESTAMPTZ;

UPDATE refunds
SET next_retry_at = updated_at,
	reconciliation_deadline = updated_at + INTERVAL '24 hours';

ALTER TABLE refunds
	ALTER COLUMN next_retry_at SET NOT NULL,
	ALTER COLUMN reconciliation_deadline SET NOT NULL,
	ADD CONSTRAINT refunds_reconciliation_time_check CHECK (reconciliation_deadline >= next_retry_at);

ALTER TABLE refunds DROP CONSTRAINT refunds_status_check;
ALTER TABLE refunds
	ADD CONSTRAINT refunds_status_check CHECK (status IN ('processing', 'success', 'failed', 'requires_review'));

CREATE INDEX refunds_reconciliation_due_idx
	ON refunds (next_retry_at, id)
	WHERE status = 'processing';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX refunds_reconciliation_due_idx;

UPDATE refunds
SET status = 'processing',
	failure_code = NULL
WHERE status = 'requires_review';

ALTER TABLE refunds DROP CONSTRAINT refunds_status_check;
ALTER TABLE refunds
	ADD CONSTRAINT refunds_status_check CHECK (status IN ('processing', 'success', 'failed'));

ALTER TABLE refunds
	DROP CONSTRAINT refunds_reconciliation_time_check,
	DROP COLUMN reconciliation_deadline,
	DROP COLUMN next_retry_at;

-- +goose StatementEnd
