CREATE TABLE poll_results (
    poll_id BIGINT NOT NULL,
    option_id BIGINT NOT NULL,
    votes_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (poll_id, option_id),

    CONSTRAINT fk_poll_results_poll
        FOREIGN KEY (poll_id)
        REFERENCES polls(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_poll_results_option
        FOREIGN KEY (option_id)
        REFERENCES poll_options(id)
        ON DELETE CASCADE,

    CONSTRAINT poll_results_votes_count_check
        CHECK (votes_count >= 0)
);