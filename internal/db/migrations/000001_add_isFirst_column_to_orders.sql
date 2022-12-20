-- +migrate Up
ALTER TABLE "orders" ADD "is_first" bool NOT NULL DEFAULT false;
-- +migrate Down
