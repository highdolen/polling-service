CREATE TABLE polls (
    id BIGSERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT polls_time_check
        CHECK (ends_at > starts_at),

    CONSTRAINT polls_type_check
        CHECK (type IN ('single', 'multiple')),

    CONSTRAINT polls_status_check
        CHECK (status IN ('draft', 'active', 'finished'))
);