-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

INSERT INTO countries (name, is_russia) VALUES
('Russia', true),
('Belarus', false),
('Kazakhstan', false),
('Germany', false),
('Turkey', false),
('France', false),
('Italy', false),
('Spain', false),
('UAE', false),
('China', false);

INSERT INTO cities (country_id, name) VALUES
(1, 'Moscow'), (1, 'Saint Petersburg'), (1, 'Kazan'),
(2, 'Minsk'), (2, 'Brest'), (2, 'Gomel'),
(3, 'Almaty'), (3, 'Astana'), (3, 'Shymkent'),
(4, 'Berlin'), (4, 'Munich'), (4, 'Hamburg'),
(5, 'Istanbul'), (5, 'Ankara'), (5, 'Antalya'),
(6, 'Paris'), (6, 'Lyon'), (6, 'Nice'),
(7, 'Rome'), (7, 'Milan'), (7, 'Venice'),
(8, 'Madrid'), (8, 'Barcelona'), (8, 'Valencia'),
(9, 'Dubai'), (9, 'Abu Dhabi'), (9, 'Sharjah'),
(10, 'Beijing'), (10, 'Shanghai'), (10, 'Guangzhou');

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DELETE FROM cities;
DELETE FROM countries;