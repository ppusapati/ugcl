-- Notification Service Database Schema
-- This schema supports the notification service for workflow and SLA events

-- Notification Types Enum
CREATE TYPE notification_type AS ENUM (
    'WORKFLOW_STATE_CHANGE',
    'WORKFLOW_ASSIGNMENT', 
    'WORKFLOW_APPROVAL_REQUEST',
    'WORKFLOW_APPROVAL_RESPONSE',
    'SLA_WARNING',
    'SLA_BREACH',
    'SLA_ESCALATION',
    'SLA_COMPLETION',
    'FORM_SUBMISSION',
    'FORM_UPDATE',
    'SYSTEM_ALERT',
    'CUSTOM'
);

-- Notification Channels Enum
CREATE TYPE notification_channel AS ENUM (
    'EMAIL',
    'SMS',
    'IN_APP',
    'PUSH',
    'WEBHOOK',
    'SLACK',
    'TEAMS'
);

-- Notification Priority Enum
CREATE TYPE notification_priority AS ENUM (
    'LOW',
    'NORMAL',
    'HIGH',
    'URGENT',
    'CRITICAL'
);

-- Notification Status Enum
CREATE TYPE notification_status AS ENUM (
    'PENDING',
    'SCHEDULED',
    'SENT',
    'DELIVERED',
    'READ',
    'FAILED',
    'CANCELLED'
);

-- SLA Event Type Enum
CREATE TYPE sla_event_type AS ENUM (
    'SLA_STARTED',
    'SLA_WARNING_THRESHOLD',
    'SLA_BREACH_IMMINENT',
    'SLA_BREACHED',
    'SLA_ESCALATED',
    'SLA_RESOLVED',
    'SLA_PAUSED',
    'SLA_RESUMED'
);

-- Core Notifications Table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    priority notification_priority DEFAULT 'NORMAL',
    recipient_id VARCHAR(255) NOT NULL,
    recipient_type VARCHAR(50) NOT NULL, -- 'user', 'role', 'email'
    subject TEXT NOT NULL,
    message TEXT NOT NULL,
    template_id UUID,
    template_data JSONB,
    status notification_status DEFAULT 'PENDING',
    source_id VARCHAR(255), -- form_instance_id, sla_instance_id, etc.
    source_type VARCHAR(50), -- 'workflow', 'sla', 'form'
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Notification Templates Table
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    subject_template TEXT NOT NULL,
    body_template TEXT NOT NULL,
    language VARCHAR(10) DEFAULT 'en',
    variables JSONB,
    active BOOLEAN DEFAULT true,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, type, channel, language)
);

-- User Notification Preferences Table
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    type notification_type NOT NULL,
    enabled_channels notification_channel[] DEFAULT ARRAY[]::notification_channel[],
    enabled BOOLEAN DEFAULT true,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, type)
);

-- Workflow State Change Notifications Table
CREATE TABLE workflow_state_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    instance_id VARCHAR(255) NOT NULL,
    form_id VARCHAR(255) NOT NULL,
    previous_state VARCHAR(100),
    current_state VARCHAR(100) NOT NULL,
    changed_by VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    assigned_role VARCHAR(255),
    form_data JSONB,
    transition_event VARCHAR(100),
    comment TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- SLA Event Notifications Table
CREATE TABLE sla_event_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    sla_instance_id VARCHAR(255) NOT NULL,
    instance_id VARCHAR(255) NOT NULL,
    sla_rule_id VARCHAR(255) NOT NULL,
    sla_rule_name VARCHAR(255) NOT NULL,
    event_type sla_event_type NOT NULL,
    state VARCHAR(100) NOT NULL,
    due_time TIMESTAMP WITH TIME ZONE,
    breach_time TIMESTAMP WITH TIME ZONE,
    assigned_to VARCHAR(255),
    severity VARCHAR(50),
    escalation_level INTEGER,
    breach_duration VARCHAR(100),
    context JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Notification Delivery Log Table
CREATE TABLE notification_delivery_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL,
    status notification_status NOT NULL,
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    response_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Notification Statistics Table (for reporting)
CREATE TABLE notification_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    type notification_type NOT NULL,
    channel notification_channel NOT NULL,
    status notification_status NOT NULL,
    count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(date, type, channel, status)
);

-- Indexes for Performance
CREATE INDEX idx_notifications_recipient ON notifications(recipient_id, recipient_type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_source ON notifications(source_id, source_type);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_at) WHERE scheduled_at IS NOT NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_read_status ON notifications(recipient_id) WHERE read_at IS NULL;

CREATE INDEX idx_templates_type_channel ON notification_templates(type, channel);
CREATE INDEX idx_templates_active ON notification_templates(active) WHERE active = true;

CREATE INDEX idx_preferences_user ON notification_preferences(user_id);
CREATE INDEX idx_preferences_type ON notification_preferences(type);

CREATE INDEX idx_workflow_notifications_instance ON workflow_state_notifications(instance_id);
CREATE INDEX idx_workflow_notifications_form ON workflow_state_notifications(form_id);
CREATE INDEX idx_workflow_notifications_state ON workflow_state_notifications(current_state);

CREATE INDEX idx_sla_notifications_instance ON sla_event_notifications(sla_instance_id);
CREATE INDEX idx_sla_notifications_rule ON sla_event_notifications(sla_rule_id);
CREATE INDEX idx_sla_notifications_event ON sla_event_notifications(event_type);

CREATE INDEX idx_delivery_log_notification ON notification_delivery_log(notification_id);
CREATE INDEX idx_delivery_log_status ON notification_delivery_log(status);

CREATE INDEX idx_stats_date_type ON notification_stats(date, type);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_notifications_updated_at 
    BEFORE UPDATE ON notifications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_templates_updated_at 
    BEFORE UPDATE ON notification_templates 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_preferences_updated_at 
    BEFORE UPDATE ON notification_preferences 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Views for Common Queries

-- Unread Notifications View
CREATE VIEW unread_notifications AS
SELECT 
    n.*,
    CASE 
        WHEN n.type IN ('SLA_WARNING', 'SLA_BREACH', 'SLA_ESCALATION') THEN 'SLA'
        WHEN n.type IN ('WORKFLOW_STATE_CHANGE', 'WORKFLOW_ASSIGNMENT', 'WORKFLOW_APPROVAL_REQUEST', 'WORKFLOW_APPROVAL_RESPONSE') THEN 'WORKFLOW'
        ELSE 'OTHER'
    END as category
FROM notifications n
WHERE n.read_at IS NULL 
    AND n.status IN ('SENT', 'DELIVERED');

-- Notification Summary View
CREATE VIEW notification_summary AS
SELECT 
    recipient_id,
    recipient_type,
    COUNT(*) as total_notifications,
    COUNT(CASE WHEN read_at IS NULL THEN 1 END) as unread_count,
    COUNT(CASE WHEN type IN ('SLA_WARNING', 'SLA_BREACH', 'SLA_ESCALATION') THEN 1 END) as sla_notifications,
    COUNT(CASE WHEN type IN ('WORKFLOW_STATE_CHANGE', 'WORKFLOW_ASSIGNMENT', 'WORKFLOW_APPROVAL_REQUEST', 'WORKFLOW_APPROVAL_RESPONSE') THEN 1 END) as workflow_notifications,
    MAX(created_at) as last_notification_at
FROM notifications
GROUP BY recipient_id, recipient_type;

-- Failed Notifications View (for retry processing)
CREATE VIEW failed_notifications AS
SELECT 
    n.*,
    dl.error_message as last_error,
    dl.attempted_at as last_attempt
FROM notifications n
LEFT JOIN notification_delivery_log dl ON n.id = dl.notification_id
WHERE n.status = 'FAILED' 
    AND (n.next_retry_at IS NULL OR n.next_retry_at <= NOW())
    AND n.retry_count < 5; -- Max retry limit

-- SLA Notification Details View
CREATE VIEW sla_notification_details AS
SELECT 
    n.id,
    n.recipient_id,
    n.subject,
    n.message,
    n.status,
    n.created_at,
    n.sent_at,
    n.read_at,
    sen.sla_rule_name,
    sen.event_type,
    sen.state,
    sen.due_time,
    sen.breach_time,
    sen.assigned_to,
    sen.severity,
    sen.escalation_level
FROM notifications n
JOIN sla_event_notifications sen ON n.id = sen.notification_id
WHERE n.type IN ('SLA_WARNING', 'SLA_BREACH', 'SLA_ESCALATION', 'SLA_COMPLETION');

-- Workflow Notification Details View
CREATE VIEW workflow_notification_details AS
SELECT 
    n.id,
    n.recipient_id,
    n.subject,
    n.message,
    n.status,
    n.created_at,
    n.sent_at,
    n.read_at,
    wsn.form_id,
    wsn.previous_state,
    wsn.current_state,
    wsn.changed_by,
    wsn.assigned_to,
    wsn.assigned_role,
    wsn.transition_event,
    wsn.changed_at
FROM notifications n
JOIN workflow_state_notifications wsn ON n.id = wsn.notification_id
WHERE n.type IN ('WORKFLOW_STATE_CHANGE', 'WORKFLOW_ASSIGNMENT', 'WORKFLOW_APPROVAL_REQUEST', 'WORKFLOW_APPROVAL_RESPONSE');
