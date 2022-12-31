CREATE TABLE "workgroups" (
  "id" bigserial PRIMARY KEY,
  "name" varchar
);

CREATE TABLE "job_titles" (
  "id" bigserial PRIMARY KEY,
  "title" varchar
);

CREATE TABLE "clearance_levels" (
  "id" bigserial PRIMARY KEY,
  "description" varchar
);

CREATE TABLE "resources" (
  "id" bigserial PRIMARY KEY,
  "first_name" varchar,
  "last_name" varchar,
  "gender_id" integer,
  "email" varchar,
  "job_title_id" integer,
  "manager_user_id" integer,
  "workgroup_id" integer,
  "clearance_level_id" integer,
  "specialties" text[],
  "certifications" text[],
  "active" boolean
);

CREATE TABLE "projects" (
  "id" bigserial PRIMARY KEY,
  "opportunity_id" varchar,
  "changepoint_id" varchar,
  "customer" varchar,
  "end_customer" varchar,
  "project_manager_user_id" int
);

CREATE TABLE "resource_requests" (
  "id" bigserial PRIMARY KEY,
  "project_id" integer,
  "job_title_id" integer,
  "skills" text[],
  "start_date" timestamptz,
  "end_date" timestamptz,
  "hours_per_week" numeric
);

CREATE TABLE "resource_assignments" (
  "id" bigserial PRIMARY KEY,
  "resource_user_id" integer,
  "resource_request_id" integer,
  "start_date" timestamptz,
  "end_date" timestamptz,
  "hours_per_week" numeric
);

CREATE TABLE "recruitment_requests" (
  "id" bigserial PRIMARY KEY,
  "resource_requests_id" integer,
  "description" varchar,
  "created_at" timestamptz,
  "filled" boolean
);

CREATE TABLE "recruitment_comments" (
  "id" bigserial PRIMARY KEY,
  "recruitment_request_id" integer,
  "comment" varchar,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "version" integer
);

ALTER TABLE "resources" ADD FOREIGN KEY ("gender_id") REFERENCES "genders" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("workgroup_id") REFERENCES "workgroups" ("id");

ALTER TABLE "users" ADD FOREIGN KEY ("clearance_level_id") REFERENCES "clearance_levels" ("id");

ALTER TABLE "projects" ADD FOREIGN KEY ("project_manager_user_id") REFERENCES "users" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_titles" ("id");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("job_title_id") REFERENCES "job_titles" ("id");

ALTER TABLE "resource_requests" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

ALTER TABLE "recruitment_requests" ADD FOREIGN KEY ("resource_requests_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "recruitment_comments" ADD FOREIGN KEY ("recruitment_request_id") REFERENCES "recruitment_requests" ("id");

ALTER TABLE "resources" ADD FOREIGN KEY ("manager_user_id") REFERENCES "resources" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource_request_id") REFERENCES "resource_requests" ("id");

ALTER TABLE "resource_assignments" ADD FOREIGN KEY ("resource_user_id") REFERENCES "resources" ("id");
