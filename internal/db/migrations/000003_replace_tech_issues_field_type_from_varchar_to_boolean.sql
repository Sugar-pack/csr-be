-- +migrate Up
BEGIN TRANSACTION;

ALTER TABLE "equipment" ADD "tech_issue_temp" bool NOT NULL DEFAULT false;

UPDATE "equipment" SET "tech_issue_temp" = (CASE "tech_issue"
    WHEN 'есть' THEN TRUE
    WHEN 'нет' THEN FALSE
    ELSE FALSE
    END);

ALTER TABLE "equipment" DROP COLUMN "tech_issue";

ALTER TABLE "equipment" ADD "tech_issue" bool NOT NULL DEFAULT false;

UPDATE "equipment"
SET "tech_issue" = "tech_issue_temp";

ALTER TABLE "equipment" DROP COLUMN "tech_issue_temp";

COMMIT;

-- +migrate Down