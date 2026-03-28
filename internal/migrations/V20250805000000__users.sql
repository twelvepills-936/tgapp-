CREATE TABLE users
(
    id          BIGSERIAL    NOT NULL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    surname     VARCHAR(255) NOT NULL,
    create_date TIMESTAMP    NOT NULL DEFAULT TIMEZONE('UTC', NOW()),
    update_date TIMESTAMP    NOT NULL DEFAULT TIMEZONE('UTC', NOW())
);
