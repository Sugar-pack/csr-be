-- +migrate Up
ALTER TABLE "registration_confirms" ALTER COLUMN "ttl" TYPE TIMESTAMPTZ;
ALTER TABLE "password_resets" ALTER COLUMN "ttl" TYPE TIMESTAMPTZ;

-- +migrate Down
ALTER TABLE "registration_confirms" ALTER COLUMN "ttl" TYPE TIMESTAMP;
ALTER TABLE "password_resets" ALTER COLUMN "ttl" TYPE TIMESTAMP;
