-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender_enum AS ENUM ('unknown', 'male', 'female');
CREATE TYPE name_type_enum AS ENUM ('name', 'surname', 'patronymic');

CREATE TABLE names (
    id SERIAL PRIMARY KEY,
    name_text TEXT NOT NULL,
    name_type name_type_enum NOT NULL,
    gender gender_enum NOT NULL DEFAULT 'unknown',
    count INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE names;
DROP TYPE gender_enum;
DROP TYPE name_type_enum;
-- +goose StatementEnd
