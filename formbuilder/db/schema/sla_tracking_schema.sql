-- =============================================================================
-- SLA Tracking Schema - Additional tables for SLA monitoring and tracking
-- =============================================================================

-- =============================================================================
-- SLA instances table - tracks SLA performance for each form instance
-- =============================================================================
CREATE TABLE sla_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    sla_rule_id UUID NOT NULL REFERENCES sla_rules(id) ON DELETE CASCADE,
    state VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    due_time TIMESTAMP WITH TIME ZONE NOT NULL,
    completion_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, completed, breached, paused
    breach_time TIMESTAMP WITH TIME ZONE,
    breach_duration INTERVAL,
    assigned_to VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_instances_status CHECK (status IN ('active', 'completed', 'breached', 'paused', 'cancelled')),
    CONSTRAINT chk_sla_instances_state_not_empty CHECK (length(trim(state)) > 0),
    CONSTRAINT chk_sla_instances_due_after_start CHECK (due_time > start_time),
    CONSTRAINT chk_sla_instances_completion_after_start CHECK (completion_time IS NULL OR completion_time >= start_time)
);

-- =============================================================================
-- SLA violations table - records when SLAs are breached
-- =============================================================================
CREATE TABLE sla_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES form_instances(id) ON DELETE CASCADE,
    sla_rule_id UUID NOT NULL REFERENCES sla_rules(id) ON DELETE CASCADE,
    violation_time TIMESTAMP WITH TIME ZONE NOT NULL,
    breach_duration INTERVAL NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, critical
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255),
    resolution_notes TEXT,
    notification_sent BOOLEAN DEFAULT false,
    escalation_triggered BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_violations_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_sla_violations_resolved_consistency CHECK (
        (resolved = false AND resolved_at IS NULL AND resolved_by IS NULL) OR
        (resolved = true AND resolved_at IS NOT NULL AND resolved_by IS NOT NULL)
    )
);

-- =============================================================================
-- SLA escalation instances table - tracks escalation executions
-- =============================================================================
CREATE TABLE sla_escalation_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    escalation_level INTEGER NOT NULL CHECK (escalation_level > 0),
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    actions_executed JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, executing, completed, failed
    error_message TEXT,
    retry_count INTEGER DEFAULT 0 CHECK (retry_count >= 0),
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_escalation_status CHECK (status IN ('pending', 'executing', 'completed', 'failed'))
);

-- =============================================================================
-- SLA notifications table - tracks all SLA-related notifications
-- =============================================================================
CREATE TABLE sla_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID REFERENCES sla_instances(id) ON DELETE CASCADE,
    sla_violation_id UUID REFERENCES sla_violations(id) ON DELETE CASCADE,
    escalation_instance_id UUID REFERENCES sla_escalation_instances(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL, -- warning, breach, escalation, reminder
    recipient VARCHAR(255) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- email, sms, push, webhook
    subject VARCHAR(500),
    message TEXT NOT NULL,
    template_used VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, sent, failed, delivered
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0 CHECK (retry_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_notifications_type CHECK (notification_type IN ('warning', 'breach', 'escalation', 'reminder', 'resolution')),
    CONSTRAINT chk_sla_notifications_channel CHECK (channel IN ('email', 'sms', 'push', 'webhook', 'in_app')),
    CONSTRAINT chk_sla_notifications_status CHECK (status IN ('pending', 'sent', 'failed', 'delivered')),
    CONSTRAINT chk_sla_notifications_recipient_not_empty CHECK (length(trim(recipient)) > 0),
    CONSTRAINT chk_sla_notifications_at_least_one_reference CHECK (
        sla_instance_id IS NOT NULL OR sla_violation_id IS NOT NULL OR escalation_instance_id IS NOT NULL
    )
);

-- =============================================================================
-- SLA metrics table - stores aggregated SLA performance metrics
-- =============================================================================
CREATE TABLE sla_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_date DATE NOT NULL,
    sla_rule_id UUID REFERENCES sla_rules(id) ON DELETE CASCADE,
    state VARCHAR(100),
    assigned_role VARCHAR(255),
    total_instances INTEGER NOT NULL DEFAULT 0 CHECK (total_instances >= 0),
    completed_on_time INTEGER NOT NULL DEFAULT 0 CHECK (completed_on_time >= 0),
    breached_instances INTEGER NOT NULL DEFAULT 0 CHECK (breached_instances >= 0),
    avg_completion_time INTERVAL,
    avg_breach_duration INTERVAL,
    compliance_percentage DECIMAL(5,2) CHECK (compliance_percentage >= 0 AND compliance_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_metrics_consistency CHECK (
        completed_on_time + breached_instances <= total_instances
    ),
    UNIQUE(metric_date, sla_rule_id, state, assigned_role)
);

-- =============================================================================
-- SLA pause logs table - tracks when SLAs are paused/resumed
-- =============================================================================
CREATE TABLE sla_pause_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sla_instance_id UUID NOT NULL REFERENCES sla_instances(id) ON DELETE CASCADE,
    action VARCHAR(10) NOT NULL, -- pause, resume
    reason TEXT,
    paused_by VARCHAR(255) NOT NULL,
    paused_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resumed_at TIMESTAMP WITH TIME ZONE,
    pause_duration INTERVAL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sla_pause_action CHECK (action IN ('pause', 'resume')),
    CONSTRAINT chk_sla_pause_consistency CHECK (
        (action = 'pause' AND resumed_at IS NULL) OR
        (action = 'resume' AND resumed_at IS NOT NULL AND resumed_at > paused_at)
    )
);

-- =============================================================================
-- Create indexes for SLA tracking tables
-- =============================================================================

-- SLA instances indexes
CREATE INDEX idx_sla_instances_instance_id ON sla_instances(instance_id);
CREATE INDEX idx_sla_instances_sla_rule_id ON sla_instances(sla_rule_id);
CREATE INDEX idx_sla_instances_state ON sla_instances(state);
CREATE INDEX idx_sla_instances_status ON sla_instances(status);
CREATE INDEX idx_sla_instances_assigned_to ON sla_instances(assigned_to);
CREATE INDEX idx_sla_instances_due_time ON sla_instances(due_time);
CREATE INDEX idx_sla_instances_active_due ON sla_instances(status, due_time) WHERE status = 'active';
CREATE INDEX idx_sla_instances_state_status ON sla_instances(state, status);

-- SLA violations indexes
CREATE INDEX idx_sla_violations_sla_instance_id ON sla_violations(sla_instance_id);
CREATE INDEX idx_sla_violations_instance_id ON sla_violations(instance_id);
CREATE INDEX idx_sla_violations_sla_rule_id ON sla_violations(sla_rule_id);
CREATE INDEX idx_sla_violations_violation_time ON sla_violations(violation_time);
CREATE INDEX idx_sla_violations_severity ON sla_violations(severity);
CREATE INDEX idx_sla_violations_resolved ON sla_violations(resolved);
CREATE INDEX idx_sla_violations_unresolved ON sla_violations(resolved, violation_time) WHERE resolved = false;

-- SLA escalation instances indexes
CREATE INDEX idx_sla_escalation_instances_sla_instance_id ON sla_escalation_instances(sla_instance_id);
CREATE INDEX idx_sla_escalation_instances_level ON sla_escalation_instances(escalation_level);
CREATE INDEX idx_sla_escalation_instances_triggered_at ON sla_escalation_instances(triggered_at);
CREATE INDEX idx_sla_escalation_instances_status ON sla_escalation_instances(status);
CREATE INDEX idx_sla_escalation_instances_retry ON sla_escalation_instances(status, next_retry_at) WHERE status = 'failed';

-- SLA notifications indexes
CREATE INDEX idx_sla_notifications_sla_instance_id ON sla_notifications(sla_instance_id);
CREATE INDEX idx_sla_notifications_sla_violation_id ON sla_notifications(sla_violation_id);
CREATE INDEX idx_sla_notifications_escalation_instance_id ON sla_notifications(escalation_instance_id);
CREATE INDEX idx_sla_notifications_type ON sla_notifications(notification_type);
CREATE INDEX idx_sla_notifications_recipient ON sla_notifications(recipient);
CREATE INDEX idx_sla_notifications_status ON sla_notifications(status);
CREATE INDEX idx_sla_notifications_channel ON sla_notifications(channel);
CREATE INDEX idx_sla_notifications_pending ON sla_notifications(status, created_at) WHERE status = 'pending';

-- SLA metrics indexes
CREATE INDEX idx_sla_metrics_date ON sla_metrics(metric_date);
CREATE INDEX idx_sla_metrics_sla_rule_id ON sla_metrics(sla_rule_id);
CREATE INDEX idx_sla_metrics_state ON sla_metrics(state);
CREATE INDEX idx_sla_metrics_role ON sla_metrics(assigned_role);
CREATE INDEX idx_sla_metrics_date_rule ON sla_metrics(metric_date, sla_rule_id);

-- SLA pause logs indexes
CREATE INDEX idx_sla_pause_logs_sla_instance_id ON sla_pause_logs(sla_instance_id);
CREATE INDEX idx_sla_pause_logs_action ON sla_pause_logs(action);
CREATE INDEX idx_sla_pause_logs_paused_by ON sla_pause_logs(paused_by);
CREATE INDEX idx_sla_pause_logs_paused_at ON sla_pause_logs(paused_at);

-- =============================================================================
-- Create views for common SLA queries
-- =============================================================================

-- Active SLA instances with time remaining
CREATE OR REPLACE VIEW active_sla_instances AS
SELECT 
    si.*,
    fi.form_id,
    fi.created_by as instance_created_by,
    sr.name as sla_rule_name,
    sr.duration as sla_duration,
    EXTRACT(EPOCH FROM (si.due_time - NOW())) / 3600 as hours_remaining,
    CASE 
        WHEN NOW() > si.due_time THEN true 
        ELSE false 
    END as is_overdue,
    EXTRACT(EPOCH FROM (NOW() - si.due_time)) / 3600 as hours_overdue
FROM sla_instances si
JOIN form_instances fi ON si.instance_id = fi.id
JOIN sla_rules sr ON si.sla_rule_id = sr.id
WHERE si.status = 'active' AND fi.deleted_at IS NULL;

-- SLA compliance summary by rule
CREATE OR REPLACE VIEW sla_compliance_summary AS
SELECT 
    sr.id as sla_rule_id,
    sr.name as sla_rule_name,
    sr.state,
    COUNT(si.id) as total_instances,
    COUNT(CASE WHEN si.status = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN si.status = 'breached' THEN 1 END) as breached_instances,
    COUNT(CASE WHEN si.status = 'active' AND NOW() > si.due_time THEN 1 END) as overdue_instances,
    ROUND(
        (COUNT(CASE WHEN si.status = 'completed' THEN 1 END) * 100.0) / 
        NULLIF(COUNT(si.id), 0), 2
    ) as compliance_percentage
FROM sla_rules sr
LEFT JOIN sla_instances si ON sr.id = si.sla_rule_id
WHERE sr.deleted_at IS NULL
GROUP BY sr.id, sr.name, sr.state;

-- Recent SLA violations
CREATE OR REPLACE VIEW recent_sla_violations AS
SELECT 
    sv.*,
    si.instance_id as si_instance_id,
    si.state,
    si.assigned_to,
    sr.name as sla_rule_name,
    fi.form_id,
    fi.created_by as instance_created_by
FROM sla_violations sv
JOIN sla_instances si ON sv.sla_instance_id = si.id
JOIN sla_rules sr ON sv.sla_rule_id = sr.id
JOIN form_instances fi ON sv.instance_id = fi.id
WHERE sv.created_at >= NOW() - INTERVAL '30 days'
ORDER BY sv.violation_time DESC;
