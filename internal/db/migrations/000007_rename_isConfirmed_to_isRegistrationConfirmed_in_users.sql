-- +migrate Up
ALTER TABLE users RENAME is_confirmed to is_registration_confirmed;

-- +migrate Down
ALTER TABLE users RENAME is_registration_confirmed to is_confirmed;