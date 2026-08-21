-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE carriers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
	code VARCHAR(3) NOT NULL,
    transport_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT carriers_transport_check
        CHECK (transport_type IN ('avia', 'rail', 'bus')),
    CONSTRAINT carriers_name_transport_unique
		UNIQUE (name, transport_type),
	CONSTRAINT carriers_code_transport_unique
		UNIQUE (code, transport_type),
	CONSTRAINT carriers_code_format_check
		CHECK (code ~ '^[A-Z0-9]{2,3}$')
);

CREATE INDEX idx_carriers_transport ON carriers (transport_type);

INSERT INTO carriers (name, code, transport_type) VALUES
('Aeroflot', 'SU', 'avia'),
('S7 Airlines', 'S7', 'avia'),
('Pobeda', 'DP', 'avia'),
('Ural Airlines', 'U6', 'avia'),
('Turkish Airlines', 'TK', 'avia'),
('Emirates', 'EK', 'avia'),
('Lufthansa', 'LH', 'avia'),
('Air France', 'AF', 'avia'),
('РЖД', 'RZD', 'rail'),
('ФПК', 'FPK', 'rail'),
('Сапсан', 'SAP', 'rail'),
('FlixBus', 'FLB', 'bus'),
('Единый Транспорт', 'ETP', 'bus'),
('Автолайн', 'ATL', 'bus'),
('Unitiki', 'UNT', 'bus');

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS carriers;
