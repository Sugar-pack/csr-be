-- +migrate Up
INSERT INTO "order_status_names" VALUES(DEFAULT,'blocked');

-- +migrate Down
DELETE FROM "order_status_names" WHERE "status"='blocked';