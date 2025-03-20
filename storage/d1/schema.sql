CREATE TABLE scores (
    id           INTEGER      PRIMARY KEY AUTOINCREMENT,
    display_name TEXT(10)     NOT NULL,
    score        INTEGER      NOT NULL,
    created_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_created_at_score ON scores (created_at, score DESC);

CREATE TABLE sessions (
    id         INTEGER      PRIMARY KEY AUTOINCREMENT,
    token      TEXT(26)     NOT NULL,
    pipe_key   TEXT(26)     NOT NULL,
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_token ON sessions (token);
