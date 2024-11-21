-- type Score struct {
-- 	ID        int       `db:"id"`
-- 	Name      string    `db:"name"`
-- 	Score     int       `db:"score"`
-- 	CreatedAt time.Time `db:"created_at"`
-- }
CREATE TABLE scores (
    id         INT          AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    score      INT          NOT NULL,
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);


-- type Session struct {
-- 	ID        int       `db:"id"`
-- 	Token     string    `db:"token"`
-- 	PipeKey   string    `db:"pipe_key"`
-- 	CreatedAt time.Time `db:"created_at"`
-- }
CREATE TABLE sessions (
    id         INT          AUTO_INCREMENT PRIMARY KEY,
    token      VARCHAR(255) NOT NULL,
    pipe_key   VARCHAR(255) NOT NULL,
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);