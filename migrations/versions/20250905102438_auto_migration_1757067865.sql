-- Create "sla_instances" table
CREATE TABLE "sla_instances" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "instance_id" uuid NOT NULL,
  "sla_rule_id" uuid NOT NULL,
  "state" character varying(100) NOT NULL,
  "start_time" timestamptz NOT NULL,
  "due_time" timestamptz NOT NULL,
  "completion_time" timestamptz NULL,
  "status" character varying(50) NOT NULL DEFAULT 'active',
  "breach_time" timestamptz NULL,
  "breach_duration" interval NULL,
  "assigned_to" character varying(255) NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_instances_instance_id_fkey" FOREIGN KEY ("instance_id") REFERENCES "form_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sla_instances_sla_rule_id_fkey" FOREIGN KEY ("sla_rule_id") REFERENCES "sla_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_instances_completion_after_start" CHECK ((completion_time IS NULL) OR (completion_time >= start_time)),
  CONSTRAINT "chk_sla_instances_due_after_start" CHECK (due_time > start_time),
  CONSTRAINT "chk_sla_instances_state_not_empty" CHECK (length(TRIM(BOTH FROM state)) > 0),
  CONSTRAINT "chk_sla_instances_status" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'completed'::character varying, 'breached'::character varying, 'paused'::character varying, 'cancelled'::character varying])::text[]))
);
-- Create index "idx_sla_instances_active_due" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_active_due" ON "sla_instances" ("status", "due_time") WHERE ((status)::text = 'active'::text);
-- Create index "idx_sla_instances_assigned_to" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_assigned_to" ON "sla_instances" ("assigned_to");
-- Create index "idx_sla_instances_due_time" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_due_time" ON "sla_instances" ("due_time");
-- Create index "idx_sla_instances_instance_id" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_instance_id" ON "sla_instances" ("instance_id");
-- Create index "idx_sla_instances_sla_rule_id" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_sla_rule_id" ON "sla_instances" ("sla_rule_id");
-- Create index "idx_sla_instances_state" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_state" ON "sla_instances" ("state");
-- Create index "idx_sla_instances_state_status" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_state_status" ON "sla_instances" ("state", "status");
-- Create index "idx_sla_instances_status" to table: "sla_instances"
CREATE INDEX "idx_sla_instances_status" ON "sla_instances" ("status");
-- Create "sla_escalation_instances" table
CREATE TABLE "sla_escalation_instances" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "sla_instance_id" uuid NOT NULL,
  "escalation_level" integer NOT NULL,
  "triggered_at" timestamptz NOT NULL,
  "actions_executed" jsonb NULL DEFAULT '[]',
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "error_message" text NULL,
  "retry_count" integer NULL DEFAULT 0,
  "next_retry_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_escalation_instances_sla_instance_id_fkey" FOREIGN KEY ("sla_instance_id") REFERENCES "sla_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_escalation_status" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'executing'::character varying, 'completed'::character varying, 'failed'::character varying])::text[])),
  CONSTRAINT "sla_escalation_instances_escalation_level_check" CHECK (escalation_level > 0),
  CONSTRAINT "sla_escalation_instances_retry_count_check" CHECK (retry_count >= 0)
);
-- Create index "idx_sla_escalation_instances_level" to table: "sla_escalation_instances"
CREATE INDEX "idx_sla_escalation_instances_level" ON "sla_escalation_instances" ("escalation_level");
-- Create index "idx_sla_escalation_instances_retry" to table: "sla_escalation_instances"
CREATE INDEX "idx_sla_escalation_instances_retry" ON "sla_escalation_instances" ("status", "next_retry_at") WHERE ((status)::text = 'failed'::text);
-- Create index "idx_sla_escalation_instances_sla_instance_id" to table: "sla_escalation_instances"
CREATE INDEX "idx_sla_escalation_instances_sla_instance_id" ON "sla_escalation_instances" ("sla_instance_id");
-- Create index "idx_sla_escalation_instances_status" to table: "sla_escalation_instances"
CREATE INDEX "idx_sla_escalation_instances_status" ON "sla_escalation_instances" ("status");
-- Create index "idx_sla_escalation_instances_triggered_at" to table: "sla_escalation_instances"
CREATE INDEX "idx_sla_escalation_instances_triggered_at" ON "sla_escalation_instances" ("triggered_at");
-- Create "sla_metrics" table
CREATE TABLE "sla_metrics" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "metric_date" date NOT NULL,
  "sla_rule_id" uuid NULL,
  "state" character varying(100) NULL,
  "assigned_role" character varying(255) NULL,
  "total_instances" integer NOT NULL DEFAULT 0,
  "completed_on_time" integer NOT NULL DEFAULT 0,
  "breached_instances" integer NOT NULL DEFAULT 0,
  "avg_completion_time" interval NULL,
  "avg_breach_duration" interval NULL,
  "compliance_percentage" numeric(5,2) NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_metrics_metric_date_sla_rule_id_state_assigned_role_key" UNIQUE ("metric_date", "sla_rule_id", "state", "assigned_role"),
  CONSTRAINT "sla_metrics_sla_rule_id_fkey" FOREIGN KEY ("sla_rule_id") REFERENCES "sla_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_metrics_consistency" CHECK ((completed_on_time + breached_instances) <= total_instances),
  CONSTRAINT "sla_metrics_breached_instances_check" CHECK (breached_instances >= 0),
  CONSTRAINT "sla_metrics_completed_on_time_check" CHECK (completed_on_time >= 0),
  CONSTRAINT "sla_metrics_compliance_percentage_check" CHECK ((compliance_percentage >= (0)::numeric) AND (compliance_percentage <= (100)::numeric)),
  CONSTRAINT "sla_metrics_total_instances_check" CHECK (total_instances >= 0)
);
-- Create index "idx_sla_metrics_date" to table: "sla_metrics"
CREATE INDEX "idx_sla_metrics_date" ON "sla_metrics" ("metric_date");
-- Create index "idx_sla_metrics_date_rule" to table: "sla_metrics"
CREATE INDEX "idx_sla_metrics_date_rule" ON "sla_metrics" ("metric_date", "sla_rule_id");
-- Create index "idx_sla_metrics_role" to table: "sla_metrics"
CREATE INDEX "idx_sla_metrics_role" ON "sla_metrics" ("assigned_role");
-- Create index "idx_sla_metrics_sla_rule_id" to table: "sla_metrics"
CREATE INDEX "idx_sla_metrics_sla_rule_id" ON "sla_metrics" ("sla_rule_id");
-- Create index "idx_sla_metrics_state" to table: "sla_metrics"
CREATE INDEX "idx_sla_metrics_state" ON "sla_metrics" ("state");
-- Create "sla_violations" table
CREATE TABLE "sla_violations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "sla_instance_id" uuid NOT NULL,
  "instance_id" uuid NOT NULL,
  "sla_rule_id" uuid NOT NULL,
  "violation_time" timestamptz NOT NULL,
  "breach_duration" interval NOT NULL,
  "severity" character varying(20) NOT NULL DEFAULT 'medium',
  "resolved" boolean NULL DEFAULT false,
  "resolved_at" timestamptz NULL,
  "resolved_by" character varying(255) NULL,
  "resolution_notes" text NULL,
  "notification_sent" boolean NULL DEFAULT false,
  "escalation_triggered" boolean NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_violations_instance_id_fkey" FOREIGN KEY ("instance_id") REFERENCES "form_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sla_violations_sla_instance_id_fkey" FOREIGN KEY ("sla_instance_id") REFERENCES "sla_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sla_violations_sla_rule_id_fkey" FOREIGN KEY ("sla_rule_id") REFERENCES "sla_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_violations_resolved_consistency" CHECK (((resolved = false) AND (resolved_at IS NULL) AND (resolved_by IS NULL)) OR ((resolved = true) AND (resolved_at IS NOT NULL) AND (resolved_by IS NOT NULL))),
  CONSTRAINT "chk_sla_violations_severity" CHECK ((severity)::text = ANY ((ARRAY['low'::character varying, 'medium'::character varying, 'high'::character varying, 'critical'::character varying])::text[]))
);
-- Create index "idx_sla_violations_instance_id" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_instance_id" ON "sla_violations" ("instance_id");
-- Create index "idx_sla_violations_resolved" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_resolved" ON "sla_violations" ("resolved");
-- Create index "idx_sla_violations_severity" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_severity" ON "sla_violations" ("severity");
-- Create index "idx_sla_violations_sla_instance_id" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_sla_instance_id" ON "sla_violations" ("sla_instance_id");
-- Create index "idx_sla_violations_sla_rule_id" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_sla_rule_id" ON "sla_violations" ("sla_rule_id");
-- Create index "idx_sla_violations_unresolved" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_unresolved" ON "sla_violations" ("resolved", "violation_time") WHERE (resolved = false);
-- Create index "idx_sla_violations_violation_time" to table: "sla_violations"
CREATE INDEX "idx_sla_violations_violation_time" ON "sla_violations" ("violation_time");
-- Create "sla_notifications" table
CREATE TABLE "sla_notifications" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "sla_instance_id" uuid NULL,
  "sla_violation_id" uuid NULL,
  "escalation_instance_id" uuid NULL,
  "notification_type" character varying(50) NOT NULL,
  "recipient" character varying(255) NOT NULL,
  "channel" character varying(20) NOT NULL,
  "subject" character varying(500) NULL,
  "message" text NOT NULL,
  "template_used" character varying(255) NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "sent_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "error_message" text NULL,
  "retry_count" integer NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_notifications_escalation_instance_id_fkey" FOREIGN KEY ("escalation_instance_id") REFERENCES "sla_escalation_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sla_notifications_sla_instance_id_fkey" FOREIGN KEY ("sla_instance_id") REFERENCES "sla_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sla_notifications_sla_violation_id_fkey" FOREIGN KEY ("sla_violation_id") REFERENCES "sla_violations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_notifications_at_least_one_reference" CHECK ((sla_instance_id IS NOT NULL) OR (sla_violation_id IS NOT NULL) OR (escalation_instance_id IS NOT NULL)),
  CONSTRAINT "chk_sla_notifications_channel" CHECK ((channel)::text = ANY ((ARRAY['email'::character varying, 'sms'::character varying, 'push'::character varying, 'webhook'::character varying, 'in_app'::character varying])::text[])),
  CONSTRAINT "chk_sla_notifications_recipient_not_empty" CHECK (length(TRIM(BOTH FROM recipient)) > 0),
  CONSTRAINT "chk_sla_notifications_status" CHECK ((status)::text = ANY ((ARRAY['pending'::character varying, 'sent'::character varying, 'failed'::character varying, 'delivered'::character varying])::text[])),
  CONSTRAINT "chk_sla_notifications_type" CHECK ((notification_type)::text = ANY ((ARRAY['warning'::character varying, 'breach'::character varying, 'escalation'::character varying, 'reminder'::character varying, 'resolution'::character varying])::text[])),
  CONSTRAINT "sla_notifications_retry_count_check" CHECK (retry_count >= 0)
);
-- Create index "idx_sla_notifications_channel" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_channel" ON "sla_notifications" ("channel");
-- Create index "idx_sla_notifications_escalation_instance_id" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_escalation_instance_id" ON "sla_notifications" ("escalation_instance_id");
-- Create index "idx_sla_notifications_pending" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_pending" ON "sla_notifications" ("status", "created_at") WHERE ((status)::text = 'pending'::text);
-- Create index "idx_sla_notifications_recipient" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_recipient" ON "sla_notifications" ("recipient");
-- Create index "idx_sla_notifications_sla_instance_id" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_sla_instance_id" ON "sla_notifications" ("sla_instance_id");
-- Create index "idx_sla_notifications_sla_violation_id" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_sla_violation_id" ON "sla_notifications" ("sla_violation_id");
-- Create index "idx_sla_notifications_status" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_status" ON "sla_notifications" ("status");
-- Create index "idx_sla_notifications_type" to table: "sla_notifications"
CREATE INDEX "idx_sla_notifications_type" ON "sla_notifications" ("notification_type");
-- Create "sla_pause_logs" table
CREATE TABLE "sla_pause_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "sla_instance_id" uuid NOT NULL,
  "action" character varying(10) NOT NULL,
  "reason" text NULL,
  "paused_by" character varying(255) NOT NULL,
  "paused_at" timestamptz NOT NULL,
  "resumed_at" timestamptz NULL,
  "pause_duration" interval NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "sla_pause_logs_sla_instance_id_fkey" FOREIGN KEY ("sla_instance_id") REFERENCES "sla_instances" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_sla_pause_action" CHECK ((action)::text = ANY ((ARRAY['pause'::character varying, 'resume'::character varying])::text[])),
  CONSTRAINT "chk_sla_pause_consistency" CHECK ((((action)::text = 'pause'::text) AND (resumed_at IS NULL)) OR (((action)::text = 'resume'::text) AND (resumed_at IS NOT NULL) AND (resumed_at > paused_at)))
);
-- Create index "idx_sla_pause_logs_action" to table: "sla_pause_logs"
CREATE INDEX "idx_sla_pause_logs_action" ON "sla_pause_logs" ("action");
-- Create index "idx_sla_pause_logs_paused_at" to table: "sla_pause_logs"
CREATE INDEX "idx_sla_pause_logs_paused_at" ON "sla_pause_logs" ("paused_at");
-- Create index "idx_sla_pause_logs_paused_by" to table: "sla_pause_logs"
CREATE INDEX "idx_sla_pause_logs_paused_by" ON "sla_pause_logs" ("paused_by");
-- Create index "idx_sla_pause_logs_sla_instance_id" to table: "sla_pause_logs"
CREATE INDEX "idx_sla_pause_logs_sla_instance_id" ON "sla_pause_logs" ("sla_instance_id");
