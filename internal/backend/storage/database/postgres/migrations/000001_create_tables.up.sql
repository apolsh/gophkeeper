BEGIN;
CREATE TABLE IF NOT EXISTS clients (
    client_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    date_last_modified BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    secret_id VARCHAR(36),
    owner BIGINT REFERENCES clients (client_id) ON DELETE CASCADE,
    name VARCHAR (100) NOT NULL,
    hash VARCHAR (64),
    description VARCHAR(256),
    enc_data bytea,
    type VARCHAR(50),
    date_last_modified BIGINT NOT NULL
);
COMMIT;
