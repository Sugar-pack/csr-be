-- +migrate Up
update users set is_registration_confirmed = true;

-- +migrate Down
update users set is_registration_confirmed = false;