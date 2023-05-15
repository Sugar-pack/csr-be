-- +migrate Up

CREATE TABLE "email_confirms"(
    "id" SERIAL PRIMARY KEY NOT NULL,
    "ttl" timestamp NOT NULL,
    "token" varchar(255) UNIQUE NOT NULL,
    "email" varchar(255) UNIQUE NOT NULL,
    "user_email_confirm" integer NULL,
    FOREIGN KEY("user_email_confirm")
    REFERENCES "users"("id") ON DELETE SET NULL
    );

-- +migrate Down
DROP TABLE IF EXISTS 'email_confirms';
