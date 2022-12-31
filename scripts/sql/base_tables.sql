CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE "job_titles" (
  "id" bigserial PRIMARY KEY,
  "title" varchar UNIQUE NOT NULL
);

CREATE TABLE "clearances" (
  "id" int PRIMARY KEY,
  "description" varchar UNIQUE NOT NULL
);

CREATE TABLE "resources" (
  "id" int PRIMARY KEY,
  "name" varchar NOT NULL,
  "job_title_id" int NOT NULL,
  "clearance_id" int NOT NULL,
  "specialties" text[],
  "certifications" text[],
  "active" boolean DEFAULT true
);

CREATE TABLE "projects" (
  "id" int PRIMARY KEY,
  "opportunity_id" varchar,
  "changepoint_id" varchar,
  "customer" varchar NOT NULL,
  "end_customer" varchar,
  "project_manager" int
);

CREATE TABLE "resource_requests" (
  "id" integer PRIMARY KEY,
  "project_id" integer NOT NULL,
  "job_title_id" integer NOT NULL,
  "skills" text[] NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date,
  "hours_per_week" numeric NOT NULL
);

CREATE TABLE "resource_assignments" (
  "id" integer PRIMARY KEY,
  "resource_id" integer,
  "resource_request_id" integer,
  "start_date" date,
  "end_date" date,
  "hours_per_week" numeric,
  "created_at" date NOT NULL DEFAULT NOW()::DATE,
  "updated_at" date,
  "version" integer NOT NULL DEFAULT 1

);

CREATE TABLE "recruitment_requests" (
  "id" integer PRIMARY KEY,
  "resource_requests_id" integer,
  "description" varchar,
  "created_at" date NOT NULL DEFAULT NOW()::DATE,
  "filled" boolean
);

CREATE TABLE "recruitment_comments" (
  "id" integer PRIMARY KEY,
  "recruitment_request_id" integer,
  "comment" varchar,
  "created_at" date NOT NULL DEFAULT NOW()::DATE,
  "updated_at" date,
  "version" integer NOT NULL DEFAULT 1
);

ALTER TABLE "resources" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_titles" ("id");

ALTER TABLE "projects" ADD FOREIGN KEY ("project_manager") REFERENCES "resources" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("clearance_id") REFERENCES "clearances" ("id");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_titles" ("id");

ALTER TABLE "recruitment_requests" ADD FOREIGN KEY ("resource_requests_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "recruitment_comments" ADD FOREIGN KEY ("recruitment_request_id") REFERENCES "recruitment_requests" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource_id") REFERENCES "resources" ("id");
