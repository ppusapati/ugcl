-- Create enum type "permission_effect"
CREATE TYPE "permission_effect" AS ENUM ('unknown', 'grant', 'forbidden');
-- Create "permission_def_groups" table
CREATE TABLE "permission_def_groups" (
  "name" text NOT NULL,
  "display_name" text NULL,
  "side" integer NOT NULL,
  "priority" integer NULL DEFAULT 0,
  "metadata" jsonb NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("name")
);
-- Create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "username" text NULL,
  "normalized_username" text NULL,
  "fullname" text NULL,
  "phone" text NULL,
  "phone_confirmed" boolean NULL DEFAULT false,
  "email" text NULL,
  "normalized_email" text NULL,
  "email_confirmed" boolean NULL DEFAULT false,
  "password" text NULL,
  "password_hash" text NOT NULL,
  "gender" integer NULL,
  "avatar" bytea NULL,
  "two_factor_enabled" boolean NULL DEFAULT false,
  "salt" bytea NULL,
  "two_factor_secret" text NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "roles_cache" text[] NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  "deleted_at" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_uuid_key" UNIQUE ("uuid")
);
-- Create index "ix_users_email" to table: "users"
CREATE INDEX "ix_users_email" ON "users" ("normalized_email");
-- Create index "ix_users_username" to table: "users"
CREATE INDEX "ix_users_username" ON "users" ("normalized_username");
-- Create "contractors" table
CREATE TABLE "contractors" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "company_name" character varying(255) NOT NULL,
  "company_type" character varying(100) NULL,
  "gst" character varying(50) NULL,
  "pan" character varying(50) NULL,
  "category" character varying(100) NULL,
  "person_id" bigint NULL,
  "associated_project" text[] NULL,
  "working_site" text[] NULL,
  "contract_start_date" timestamp NULL,
  "contract_end_date" timestamp NULL,
  "status" character varying(50) NULL DEFAULT 'active',
  "metadata" jsonb NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamp NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "contractors_pan_key" UNIQUE ("pan"),
  CONSTRAINT "contractors_uuid_key" UNIQUE ("uuid"),
  CONSTRAINT "contractors_person_id_fkey" FOREIGN KEY ("person_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_contract_dates" CHECK ((contract_end_date IS NULL) OR (contract_end_date > contract_start_date))
);
-- Create index "idx_contractors_category" to table: "contractors"
CREATE INDEX "idx_contractors_category" ON "contractors" ("category") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_company" to table: "contractors"
CREATE INDEX "idx_contractors_company" ON "contractors" ("company_name") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_gst" to table: "contractors"
CREATE INDEX "idx_contractors_gst" ON "contractors" ("gst") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_pan" to table: "contractors"
CREATE INDEX "idx_contractors_pan" ON "contractors" ("pan") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_person_id" to table: "contractors"
CREATE INDEX "idx_contractors_person_id" ON "contractors" ("person_id") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_projects" to table: "contractors"
CREATE INDEX "idx_contractors_projects" ON "contractors" USING gin ("associated_project") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_sites" to table: "contractors"
CREATE INDEX "idx_contractors_sites" ON "contractors" USING gin ("working_site") WHERE (deleted_at IS NULL);
-- Create index "idx_contractors_status" to table: "contractors"
CREATE INDEX "idx_contractors_status" ON "contractors" ("status") WHERE (deleted_at IS NULL);
-- Create "dairy_sites" table
CREATE TABLE "dairy_sites" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name_of_site" text NOT NULL,
  "todays_work" text NULL,
  "employee_id" bigint NOT NULL,
  "latitude" double precision NULL,
  "longitude" double precision NULL,
  "submitted_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dairy_sites_uuid_key" UNIQUE ("uuid"),
  CONSTRAINT "dairy_sites_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_dairy_sites_employee_id" to table: "dairy_sites"
CREATE INDEX "idx_dairy_sites_employee_id" ON "dairy_sites" ("employee_id");
-- Create index "idx_dairy_sites_not_deleted" to table: "dairy_sites"
CREATE INDEX "idx_dairy_sites_not_deleted" ON "dairy_sites" ("deleted_at") WHERE (deleted_at IS NULL);
-- Create "permission_defs" table
CREATE TABLE "permission_defs" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "description" text NULL,
  "side" integer NULL,
  "metadata" jsonb NULL,
  "namespace" text NOT NULL,
  "resource" text NOT NULL,
  "action" text NOT NULL,
  "scope" text NOT NULL DEFAULT '*',
  "version" integer NOT NULL DEFAULT 1,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "permission_defs_name_key" UNIQUE ("name"),
  CONSTRAINT "permission_defs_uuid_key" UNIQUE ("uuid")
);
-- Create index "ix_permdef_lookup" to table: "permission_defs"
CREATE INDEX "ix_permdef_lookup" ON "permission_defs" ("namespace", "resource", "scope");
-- Create "roles" table
CREATE TABLE "roles" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "parent_id" bigint NULL,
  "name" text NOT NULL,
  "is_preserved" boolean NOT NULL DEFAULT false,
  "metadata" jsonb NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "roles_name_key" UNIQUE ("name"),
  CONSTRAINT "roles_uuid_key" UNIQUE ("uuid"),
  CONSTRAINT "roles_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "roles" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);
-- Create index "ix_roles_parent" to table: "roles"
CREATE INDEX "ix_roles_parent" ON "roles" ("parent_id");
-- Create "role_permission_defs" table
CREATE TABLE "role_permission_defs" (
  "role_id" bigint NOT NULL,
  "permission_def_name" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("role_id", "permission_def_name"),
  CONSTRAINT "role_permission_defs_permission_def_name_fkey" FOREIGN KEY ("permission_def_name") REFERENCES "permission_defs" ("name") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "role_permission_defs_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "permissions" table
CREATE TABLE "permissions" (
  "id" bigserial NOT NULL,
  "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
  "namespace" text NOT NULL,
  "resource" text NOT NULL,
  "action" text NOT NULL,
  "subject" text NOT NULL,
  "effect" "permission_effect" NOT NULL DEFAULT 'grant',
  "def_name" text NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "permissions_namespace_resource_action_subject_key" UNIQUE ("namespace", "resource", "action", "subject"),
  CONSTRAINT "permissions_uuid_key" UNIQUE ("uuid")
);
-- Create index "ix_permissions_match" to table: "permissions"
CREATE INDEX "ix_permissions_match" ON "permissions" ("namespace", "resource", "action", "subject");
-- Create "role_permissions" table
CREATE TABLE "role_permissions" (
  "role_id" bigint NOT NULL,
  "permission_id" bigint NOT NULL,
  "granted" boolean NULL DEFAULT true,
  "created_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("role_id", "permission_id"),
  CONSTRAINT "role_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "role_permissions_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "user_permissions" table
CREATE TABLE "user_permissions" (
  "user_id" bigint NOT NULL,
  "permission_id" bigint NOT NULL,
  "granted" boolean NULL DEFAULT true,
  PRIMARY KEY ("user_id", "permission_id"),
  CONSTRAINT "user_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_permissions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "user_roles" table
CREATE TABLE "user_roles" (
  "user_id" bigint NOT NULL,
  "role_id" bigint NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id", "role_id"),
  CONSTRAINT "user_roles_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_roles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
