-- +migrate Up
BEGIN TRANSACTION;

ALTER TABLE "tokens" ALTER COLUMN access_token TYPE varchar(400);

COMMIT;

-- +migrate Down