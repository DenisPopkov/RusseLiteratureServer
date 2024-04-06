CREATE TABLE IF NOT EXISTS users
(
    id        INTEGER PRIMARY KEY,
    phone     TEXT    NOT NULL UNIQUE,
    pass_hash BLOB    NOT NULL,
    name      TEXT    NOT NULL,
    image     TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_phone ON users (phone);

CREATE TABLE IF NOT EXISTS apps
(
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);
