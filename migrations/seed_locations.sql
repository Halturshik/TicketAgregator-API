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

UPDATE cities
SET is_air_hub = true
WHERE name IN ('Moscow', 'Saint Petersburg', 'Istanbul', 'Berlin', 'Paris', 'Dubai', 'Beijing', 'Almaty');

UPDATE cities SET latitude = 55.7558, longitude = 37.6173 WHERE name = 'Moscow';
UPDATE cities SET latitude = 59.9343, longitude = 30.3351 WHERE name = 'Saint Petersburg';
UPDATE cities SET latitude = 43.2389, longitude = 76.8897 WHERE name = 'Almaty';
UPDATE cities SET latitude = 52.5200, longitude = 13.4050 WHERE name = 'Berlin';
UPDATE cities SET latitude = 41.0082, longitude = 28.9784 WHERE name = 'Istanbul';
UPDATE cities SET latitude = 48.8566, longitude = 2.3522 WHERE name = 'Paris';
UPDATE cities SET latitude = 25.2048, longitude = 55.2708 WHERE name = 'Dubai';
UPDATE cities SET latitude = 39.9042, longitude = 116.4074 WHERE name = 'Beijing';

UPDATE countries SET neighbor_country_ids = ARRAY[2,3] WHERE id = 1;
UPDATE countries SET neighbor_country_ids = ARRAY[1] WHERE id = 2;
UPDATE countries SET neighbor_country_ids = ARRAY[1,10] WHERE id = 3;
UPDATE countries SET neighbor_country_ids = ARRAY[6,7] WHERE id = 4;
UPDATE countries SET neighbor_country_ids = ARRAY[]::INT[] WHERE id = 5;
UPDATE countries SET neighbor_country_ids = ARRAY[4,7,8] WHERE id = 6;
UPDATE countries SET neighbor_country_ids = ARRAY[4,6] WHERE id = 7;
UPDATE countries SET neighbor_country_ids = ARRAY[6] WHERE id = 8;
UPDATE countries SET neighbor_country_ids = ARRAY[]::INT[] WHERE id = 9;
UPDATE countries SET neighbor_country_ids = ARRAY[3] WHERE id = 10;

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DELETE FROM cities;
DELETE FROM countries;
