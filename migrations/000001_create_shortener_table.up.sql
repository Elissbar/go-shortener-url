-- Создание таблицы
CREATE TABLE shorted_links (
    id SERIAL PRIMARY KEY,
    token VARCHAR(20) NOT NULL,
    url VARCHAR(255) NOT NULL
);

CREATE INDEX idx_token ON shorted_links(token);