-- +migrate Up
ALTER TABLE "users" RENAME "is_blocked" to "is_readonly";

-- +migrate Down
ALTER TABLE "users" RENAME "is_readonly" to "is_blocked";