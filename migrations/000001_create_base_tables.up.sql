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
  "workgroup_id" bigserail PRIMARY KEY,
  "workgroup_name" varchar,
  "description" varchar
);

CREATE TABLE "job_title" (
  "title_id" bigserial PRIMARY KEY,
  "title" varchar,
  "description" varchar
);

CREATE TABLE "resource" (
  "employee_id" integer PRIMARY KEY,
  "name" varchar,
  "email" varchar,
  "job_title_id" integer,
  "manager_id" integer,
  "workgroup_id" integer,
  "clearance_level" clearance,
  "specialties" text[],
  "certifications" text[],
  "active" boolean
);

CREATE TABLE "project_status" (
  "status_id" bigserial PRIMARY KEY,
  "status" varchar,
  "description" varchar
);

CREATE TABLE "project" (
  "opportunity_id" varchar PRIMARY KEY,
  "changepoint_id" varchar,
  "name" varchar,
  "revenue_type" revenue_type,
  "customer" varchar,
  "end_customer" varchar,
  "project_manager_id" int,
  "status" varchar
);

CREATE TABLE "resource_request" (
  "request_id" bigserial PRIMARY KEY,
  "opportunity_id" integer,
  "job_title_id" integer,
  "total_hours" numeric,
  "skills" text[],
  "start_date" date,
  "hours_per_week" numeric,
  "status" varchar,
  "created_at" timestamp,
  "updated_at" timestamp,
  "version" integer
);

CREATE TABLE "resource_request_comment" (
  "comment_id" bigserial PRIMARY KEY,
  "request_id" integer,
  "comment" varchar,
  "created_at" timestamp,
  "updated_at" timestamp,
  "version" integer
);

CREATE TABLE "resource_assignment" (
  "assignment_id" bigserial PRIMARY KEY,
  "employee_id" integer,
  "resource_request_id" integer,
  "start_date" date,
  "end_date" date,
  "hours_per_week" numeric
);

CREATE TABLE "new_hire" (
  "requirement_id" bigserial PRIMARY KEY,
  "resource_request_id" integer,
  "description" varchar,
  "created_at" timestamp,
  "filled" boolean,
  "status" varchar
);

CREATE TABLE "new_hire_update" (
  "update_id" bigserial PRIMARY KEY,
  "requirement_id" integer,
  "comment" varchar,
  "created_at" timestamp,
  "updated_at" timestamp,
  "version" integer
);

ALTER TABLE "resource" ADD FOREIGN KEY ("workgroup_id") REFERENCES "workgroup" ("workgroup_id");

ALTER TABLE "resource" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_title" ("title_id");

ALTER TABLE "resource_request" ADD FOREIGN KEY ("opportunity_id") REFERENCES "project" ("opportunity_id");

ALTER TABLE "project" ADD FOREIGN KEY ("status") REFERENCES "project_status" ("status");

ALTER TABLE "new_hire" ADD FOREIGN KEY ("requirement_id") REFERENCES "resource_request" ("request_id");

ALTER TABLE "new_hire_update" ADD FOREIGN KEY ("requirement_id") REFERENCES "new_hire" ("requirement_id");

ALTER TABLE "resource_assignment" ADD FOREIGN KEY ("employee_id") REFERENCES "resource" ("employee_id");

ALTER TABLE "resource_assignment" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_request" ("opportunity_id");

ALTER TABLE "resource" ADD FOREIGN KEY ("manager_id") REFERENCES "resource" ("employee_id");

ALTER TABLE "resource_request" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_title" ("title_id");

ALTER TABLE "resource_request_comment" ADD FOREIGN KEY ("request_id") REFERENCES "resource_request" ("request_id");
