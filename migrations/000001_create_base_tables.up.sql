CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE "clearance" AS ENUM (
  'None',
  'Baseline',
  'NV1',
  'NV2',
  'TSPV'
);

CREATE TYPE "revenue_type" AS ENUM (
  'Fixed Fee',
  'T&M'
);

CREATE TABLE "workgroup" (
  "workgroup_id" bigserial PRIMARY KEY,
  "workgroup_name" varchar NOT NULL UNIQUE,
  "description" varchar
);

CREATE TABLE "job_title" (
  "title_id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL UNIQUE,
  "description" varchar
);

CREATE TABLE "resource" (
  "employee_id" integer PRIMARY KEY,
  "name" varchar NOT NULL,
  "email" varchar NOT NULL UNIQUE,
  "job_title_id" integer NOT NULL,
  "manager_id" integer NOT NULL,
  "workgroup_id" integer NOT NULL,
  "clearance" clearance,
  "specialties" text[],
  "certifications" text[],
  "active" boolean
);

CREATE TABLE "project_status" (
  "status_id" bigserial PRIMARY KEY,
  "status" varchar NOT NULL UNIQUE,
  "description" varchar
);

CREATE TABLE "project" (
  "opportunity_id" varchar PRIMARY KEY,
  "changepoint_id" varchar UNIQUE,
  "name" varchar NOT NULL,
  "revenue_type" revenue_type,
  "customer" varchar NOT NULL,
  "end_customer" varchar,
  "project_manager_id" int,
  "status_id" integer
);

CREATE TABLE "resource_request" (
  "request_id" bigserial PRIMARY KEY,
  "opportunity_id" varchar NOT NULL,
  "job_title_id" integer NOT NULL,
  "total_hours" numeric,
  "skills" text[] NOT NULL,
  "start_date" date,
  "hours_per_week" numeric,
  "status" varchar,
  "created_at" timestamp NOT NULL DEFAULT current_timestamp,
  "updated_at" timestamp NOT NULL DEFAULT current_timestamp,
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "resource_request_comment" (
  "comment_id" bigserial PRIMARY KEY,
  "request_id" integer NOT NULL,
  "comment" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT current_timestamp,
  "updated_at" timestamp NOT NULL DEFAULT current_timestamp,
  "version" integer NOT NULL DEFAULT 1
);

CREATE TABLE "resource_assignment" (
  "assignment_id" bigserial PRIMARY KEY,
  "resource_request_id" integer NOT NULL,
  "employee_id" integer NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date,
  "hours_per_week" numeric NOT NULL
);

CREATE TABLE "new_hire" (
  "requirement_id" bigserial PRIMARY KEY,
  "resource_request_id" integer NOT NULL,
  "description" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT current_timestamp,
  "filled" boolean NOT NULL DEFAULT 'f',
  "status" varchar
);

CREATE TABLE "new_hire_update" (
  "update_id" bigserial PRIMARY KEY,
  "requirement_id" integer,
  "comment" varchar,
  "created_at" timestamp NOT NULL DEFAULT current_timestamp,
  "updated_at" timestamp NOT NULL DEFAULT current_timestamp,
  "version" integer NOT NULL DEFAULT 1
);

ALTER TABLE "resource" ADD FOREIGN KEY ("workgroup_id") REFERENCES "workgroup" ("workgroup_id");

ALTER TABLE "resource" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_title" ("title_id");

ALTER TABLE "resource" ADD FOREIGN KEY ("manager_id") REFERENCES "resource" ("employee_id");

ALTER TABLE "project" ADD FOREIGN KEY ("status_id") REFERENCES "project_status" ("status_id");

ALTER TABLE "resource_request" ADD FOREIGN KEY ("opportunity_id") REFERENCES "project" ("opportunity_id");

ALTER TABLE "resource_request" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_title" ("title_id");

ALTER TABLE "resource_request_comment" ADD FOREIGN KEY ("request_id") REFERENCES "resource_request" ("request_id");

ALTER TABLE "new_hire" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_request" ("request_id");

ALTER TABLE "new_hire_update" ADD FOREIGN KEY ("requirement_id") REFERENCES "new_hire" ("requirement_id");

ALTER TABLE "resource_assignment" ADD FOREIGN KEY ("employee_id") REFERENCES "resource" ("employee_id");

ALTER TABLE "resource_assignment" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_request" ("request_id");
