-- +goose Up
-- +goose StatementBegin
CREATE TABLE short_url (
    id SERIAL PRIMARY KEY,
    link VARCHAR(251) NOT NULL,
    times_clicked INT DEFAULT 0,
    exp_time_minutes INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE short_url;
-- +goose StatementEnd
