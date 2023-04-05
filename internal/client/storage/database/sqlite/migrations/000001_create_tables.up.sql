CREATE TABLE IF NOT EXISTS clients (
    client_id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    date_last_modified INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    secret_id TEXT,
    owner INTEGER REFERENCES clients (client_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    hash TEXT NOT NULL,
    description TEXT,
    enc_data BLOB,
    type TEXT,
    date_last_modified INTEGER NOT NULL
);

CREATE UNIQUE INDEX unique_secret_name ON secrets (name);
