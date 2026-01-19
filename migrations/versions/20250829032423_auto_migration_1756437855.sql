-- Create "escalations" table
CREATE TABLE "escalations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NOT NULL,
  "trigger_condition" text NULL,
  "from_states" jsonb NULL DEFAULT '[]',
  "to_state" character varying(100) NOT NULL,
  "notify_roles" jsonb NULL DEFAULT '[]',
  "escalation_message" text NULL,
  "auto_escalate" boolean NULL DEFAULT false,
  "after_duration" jsonb NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "escalations_name_check" CHECK (length(TRIM(BOTH FROM name)) > 0)
);
-- Create index "idx_escalations_auto_escalate" to table: "escalations"
CREATE INDEX "idx_escalations_auto_escalate" ON "escalations" ("auto_escalate");
-- Create index "idx_escalations_deleted_at" to table: "escalations"
CREATE INDEX "idx_escalations_deleted_at" ON "escalations" ("deleted_at");
-- Create index "idx_escalations_from_states" to table: "escalations"
CREATE INDEX "idx_escalations_from_states" ON "escalations" USING gin ("from_states");
-- Create "sla_rules" table
CREATE TABLE "sla_rules" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying(255) NOT NULL,
  "state" character varying(100) NOT NULL,
  "duration" jsonb NOT NULL,
  "escalation_levels" jsonb NULL DEFAULT '[]',
  "active" boolean NULL DEFAULT true,
  "applicable_roles" jsonb NULL DEFAULT '[]',
  "condition" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_rules_name_check" CHECK (length(TRIM(BOTH FROM name)) > 0)
);
-- Create index "idx_sla_rules_active" to table: "sla_rules"
CREATE INDEX "idx_sla_rules_active" ON "sla_rules" ("active");
-- Create index "idx_sla_rules_deleted_at" to table: "sla_rules"
CREATE INDEX "idx_sla_rules_deleted_at" ON "sla_rules" ("deleted_at");
-- Create index "idx_sla_rules_state" to table: "sla_rules"
CREATE INDEX "idx_sla_rules_state" ON "sla_rules" ("state");
-- Create "workflows" table
CREATE TABLE "workflows" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "initial_state" character varying(100) NOT NULL,
  "states" jsonb NOT NULL DEFAULT '[]',
  "global_transitions" jsonb NULL DEFAULT '[]',
  "metadata" jsonb NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_workflows_deleted_at" to table: "workflows"
CREATE INDEX "idx_workflows_deleted_at" ON "workflows" ("deleted_at");
-- Create index "idx_workflows_initial_state" to table: "workflows"
CREATE INDEX "idx_workflows_initial_state" ON "workflows" ("initial_state");
-- Create "forms" table
CREATE TABLE "forms" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "title" character varying(255) NOT NULL,
  "description" text NULL,
  "version" character varying(50) NOT NULL DEFAULT '1.0.0',
  "created_by" character varying(255) NOT NULL,
  "allowed_roles" jsonb NULL DEFAULT '[]',
  "audit" boolean NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "table_name" character varying(255) NULL,
  "schema_version" integer NULL DEFAULT 1,
  "core_fields" jsonb NULL DEFAULT '[]',
  "steps" jsonb NOT NULL DEFAULT '[]',
  "dependencies" jsonb NULL DEFAULT '[]',
  "cross_field_validations" jsonb NULL DEFAULT '[]',
  "workflow_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_forms_workflow_id" FOREIGN KEY ("workflow_id") REFERENCES "workflows" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "chk_forms_schema_version_positive" CHECK (schema_version > 0),
  CONSTRAINT "chk_forms_version_format" CHECK ((version)::text ~ '^[0-9]+\.[0-9]+\.[0-9]+$'::text)
);
-- Create index "idx_forms_created_by" to table: "forms"
CREATE INDEX "idx_forms_created_by" ON "forms" ("created_by");
-- Create index "idx_forms_created_by_created_at" to table: "forms"
CREATE INDEX "idx_forms_created_by_created_at" ON "forms" ("created_by", "created_at" DESC);
-- Create index "idx_forms_deleted_at" to table: "forms"
CREATE INDEX "idx_forms_deleted_at" ON "forms" ("deleted_at");
-- Create index "idx_forms_title" to table: "forms"
CREATE INDEX "idx_forms_title" ON "forms" ("title");
-- Create index "idx_forms_version" to table: "forms"
CREATE INDEX "idx_forms_version" ON "forms" ("version");
-- Create index "idx_forms_workflow_id" to table: "forms"
CREATE INDEX "idx_forms_workflow_id" ON "forms" ("workflow_id");
-- Create "form_instances" table
CREATE TABLE "form_instances" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "form_id" uuid NOT NULL,
  "current_state" character varying(100) NOT NULL DEFAULT 'draft',
  "field_values" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "created_by" character varying(255) NOT NULL,
  "assigned_to" character varying(255) NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "form_instances_form_id_fkey" FOREIGN KEY ("form_id") REFERENCES "forms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_instances_current_state_not_empty" CHECK (length(TRIM(BOTH FROM current_state)) > 0)
);
-- Create index "idx_form_instances_assigned_to" to table: "form_instances"
CREATE INDEX "idx_form_instances_assigned_to" ON "form_instances" ("assigned_to");
-- Create index "idx_form_instances_created_by" to table: "form_instances"
CREATE INDEX "idx_form_instances_created_by" ON "form_instances" ("created_by");
-- Create index "idx_form_instances_current_state" to table: "form_instances"
CREATE INDEX "idx_form_instances_current_state" ON "form_instances" ("current_state");
-- Create index "idx_form_instances_deleted_at" to table: "form_instances"
CREATE INDEX "idx_form_instances_deleted_at" ON "form_instances" ("deleted_at");
-- Create index "idx_form_instances_form_id" to table: "form_instances"
CREATE INDEX "idx_form_instances_form_id" ON "form_instances" ("form_id");
-- Create index "idx_form_instances_form_state" to table: "form_instances"
CREATE INDEX "idx_form_instances_form_state" ON "form_instances" ("form_id", "current_state");
-- Create index "idx_form_instances_state_created" to table: "form_instances"
CREATE INDEX "idx_form_instances_state_created" ON "form_instances" ("current_state", "created_at");
-- Create "attachments" table
CREATE TABLE "attachments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "instance_id" uuid NOT NULL,
  "field_id" character varying(255) NOT NULL,
  "filename" character varying(255) NOT NULL,
  "mime_type" character varying(100) NULL,
  "size" bigint NULL,
  "storage_path" text NOT NULL,
  "uploaded_at" timestamptz NOT NULL DEFAULT now(),
  "uploaded_by" character varying(255) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "attachments_instance_id_fkey" FOREIGN KEY ("instance_id") REFERENCES "form_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "attachments_size_check" CHECK (size >= 0),
  CONSTRAINT "chk_attachments_filename_not_empty" CHECK (length(TRIM(BOTH FROM filename)) > 0)
);
-- Create index "idx_attachments_field_id" to table: "attachments"
CREATE INDEX "idx_attachments_field_id" ON "attachments" ("field_id");
-- Create index "idx_attachments_instance_id" to table: "attachments"
CREATE INDEX "idx_attachments_instance_id" ON "attachments" ("instance_id");
-- Create "audit_logs" table
CREATE TABLE "audit_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "instance_id" uuid NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "action" character varying(100) NOT NULL,
  "from_state" character varying(100) NULL,
  "to_state" character varying(100) NULL,
  "changes" jsonb NULL DEFAULT '{}',
  "timestamp" timestamptz NOT NULL DEFAULT now(),
  "ip_address" inet NULL,
  "user_agent" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "audit_logs_instance_id_fkey" FOREIGN KEY ("instance_id") REFERENCES "form_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_audit_action_not_empty" CHECK (length(TRIM(BOTH FROM action)) > 0)
);
-- Create index "idx_audit_logs_action" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_action" ON "audit_logs" ("action");
-- Create index "idx_audit_logs_instance_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_instance_id" ON "audit_logs" ("instance_id");
-- Create index "idx_audit_logs_instance_timestamp" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_instance_timestamp" ON "audit_logs" ("instance_id", "timestamp" DESC);
-- Create index "idx_audit_logs_timestamp" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_timestamp" ON "audit_logs" ("timestamp");
-- Create index "idx_audit_logs_user_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_user_id" ON "audit_logs" ("user_id");
-- Create "comments" table
CREATE TABLE "comments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "instance_id" uuid NOT NULL,
  "user_id" character varying(255) NOT NULL,
  "text" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "internal" boolean NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "comments_instance_id_fkey" FOREIGN KEY ("instance_id") REFERENCES "form_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "comments_text_check" CHECK (length(TRIM(BOTH FROM text)) > 0)
);
-- Create index "idx_comments_created_at" to table: "comments"
CREATE INDEX "idx_comments_created_at" ON "comments" ("created_at");
-- Create index "idx_comments_instance_id" to table: "comments"
CREATE INDEX "idx_comments_instance_id" ON "comments" ("instance_id");
