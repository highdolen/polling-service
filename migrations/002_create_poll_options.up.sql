CREATE TABLE poll_options (
    id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL,
    text TEXT NOT NULL,

    CONSTRAINT fk_poll_options_poll
        FOREIGN KEY (poll_id)
        REFERENCES polls(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_poll_options_poll_id
    ON poll_options(poll_id);