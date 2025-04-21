-- +goose Up
-- +goose StatementBegin
ALTER TABLE short_url
ADD COLUMN short_code varchar(10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE short_url
DROP COLUMN IF EXISTS short_code;
-- +goose StatementEnd
