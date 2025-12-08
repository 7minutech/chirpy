-- +goose Up
ALTER TABLE users 
ADD column is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE users
DROP column is_chirpy_red;