-- +migrate Up
ALTER TABLE "equipment" ADD "is_archived" bool NOT NULL DEFAULT false;

-- +migrate Down
ALTER TABLE "equipment" DROP COLUMN "is_archived";