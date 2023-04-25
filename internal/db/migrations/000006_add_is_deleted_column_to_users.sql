-- +migrate Up
ALTER TABLE "users" ADD "is_deleted" bool NOT NULL DEFAULT false;

-- +migrate Down
ALTER TABLE "users" DROP COLUMN "is_deleted";