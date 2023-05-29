-- +migrate Up
INSERT INTO "equipment_status_names" VALUES(DEFAULT,'archived');

-- +migrate Down
DELETE FROM "equipment_status_names" WHERE "name"='archived';
