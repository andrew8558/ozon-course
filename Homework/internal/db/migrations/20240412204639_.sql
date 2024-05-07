-- +goose Up
-- +goose StatementBegin
CREATE TABLE pickup_points(
                              id BIGSERIAL primary key not null,
                              name varchar(256) not null,
                              address varchar(256) not null,
                              contact_details varchar(256) not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pickup_points;
-- +goose StatementEnd
