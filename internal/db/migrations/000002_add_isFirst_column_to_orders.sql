-- +migrate Up
ALTER TABLE "orders" ADD "is_first" bool NOT NULL DEFAULT false;

-- +migrate Down
ALTER TABLE "orders" DROP COLUMN "is_first";