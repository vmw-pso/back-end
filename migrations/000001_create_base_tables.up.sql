CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE "workgroups" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "description" varchar
);

CREATE TABLE "job_titles" (
  "title" varchar PRIMARY KEY
);

CREATE TABLE "clearances" (
  "level" varchar PRIMARY KEY
);

CREATE TABLE "resources" (
  "id" integer PRIMARY KEY,
  "name" varchar NOT NULL,
  "email" varchar NOT NULL,
  "job_title" varchar NOT NULL,
  "manager_id" integer NOT NULL,
  "workgroup_id" integer NOT NULL,
  "clearance" varchar,
  "specialties" text[],
  "certifications" text[],
  "active" boolean NOT NULL DEFAULT 't'
);

CREATE TABLE "project_statuses" (
  "status" varchar PRIMARY KEY
);

CREATE TABLE "projects" (
  "id" bigserial PRIMARY KEY,
  "opportunity_id" varchar,
  "changepoint_id" varchar,
  "customer" varchar NOT NULL,
  "end_customer" varchar,
  "project_manager_id" int NOT NULL,
  "status" varchar NOT NULL
);

CREATE TABLE "resource_requests" (
  "id" bigserial PRIMARY KEY,
  "project_id" integer NOT NULL,
  "job_title" varchar NOT NULL,
  "skills" text[] NOT NULL,
  "start_date" date,
  "end_date" date,
  "hours_per_week" numeric,
  "status" varchar,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "resource_assignments" (
  "id" bigserial PRIMARY KEY,
  "resource__id" integer NOT NULL,
  "resource_request_id" integer NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date NOT NULL,
  "hours_per_week" numeric NOT NULL
);

CREATE TABLE "recruitment_requests" (
  "id" bigserial PRIMARY KEY,
  "resource_requests_id" integer NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "filled" boolean,
  "status" varchar
);

CREATE TABLE "recruitment_comments" (
  "id" bigserial PRIMARY KEY,
  "recruitment_request_id" integer NOT NULL,
  "comment" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),
  "version" integer NOT NULL DEFAULT 1
);

ALTER TABLE "resources" ADD FOREIGN KEY ("workgroup_id") REFERENCES "workgroups" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("clearance") REFERENCES "clearances" ("level");

ALTER TABLE "resources" ADD FOREIGN KEY ("job_title") REFERENCES "job_titles" ("title");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

ALTER TABLE "projects" ADD FOREIGN KEY ("status") REFERENCES "project_statuses" ("status");

ALTER TABLE "recruitment_requests" ADD FOREIGN KEY ("resource_requests_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "recruitment_comments" ADD FOREIGN KEY ("recruitment_request_id") REFERENCES "recruitment_requests" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource__id") REFERENCES "resources" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("manager_id") REFERENCES "resources" ("id");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("job_title") REFERENCES "job_titles" ("title");
