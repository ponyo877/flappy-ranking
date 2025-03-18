CREATE TABLE flappy.scores (
    id           INT         AUTO_INCREMENT PRIMARY KEY,
    display_name VARCHAR(10) NOT NULL,
    score        INT         NOT NULL,
    created_at   TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_created_at_acore (created_at, score DESC)
);

CREATE TABLE flappy.sessions (
    id         INT         AUTO_INCREMENT PRIMARY KEY,
    token      VARCHAR(26) NOT NULL,
    pipe_key   VARCHAR(26) NOT NULL,
    created_at TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token)
);
