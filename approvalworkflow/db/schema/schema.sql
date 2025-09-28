-- =============================================================================
-- ADVANCED APPROVAL WORKFLOWS DATABASE SCHEMA
-- =============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- APPROVAL WORKFLOWS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_workflows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    entity_type VARCHAR(100) NOT NULL, -- 'purchase_order', 'expense', 'leave_request', etc.
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,

    -- Workflow configuration (JSON)
    steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    escalations JSONB NOT NULL DEFAULT '[]'::jsonb,
    notifications JSONB NOT NULL DEFAULT '[]'::jsonb,
    conditional_logic JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Metadata
    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Constraints
    UNIQUE(entity_type, version),
    UNIQUE(entity_type, is_default) WHERE is_default = true
);

-- =============================================================================
-- APPROVAL INSTANCES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES approval_workflows(id) ON DELETE RESTRICT,
    entity_type VARCHAR(100) NOT NULL,
    entity_id VARCHAR(255) NOT NULL, -- Reference to the actual entity being approved
    requested_by UUID NOT NULL,
    current_step VARCHAR(100),
    status VARCHAR(50) DEFAULT 'PENDING',
    priority VARCHAR(20) DEFAULT 'MEDIUM',

    -- Request data
    request_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    comments TEXT,
    attachment_urls TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Tracking timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    escalated_at TIMESTAMP WITH TIME ZONE,
    escalation_level INTEGER DEFAULT 0,

    -- SLA tracking
    sla_start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sla_end_time TIMESTAMP WITH TIME ZONE,
    sla_breached BOOLEAN DEFAULT false,

    metadata JSONB DEFAULT '{}'::jsonb,

    -- Indexes
    CONSTRAINT chk_status CHECK (status IN ('PENDING', 'IN_PROGRESS', 'APPROVED', 'REJECTED', 'CANCELLED', 'ESCALATED', 'EXPIRED')),
    CONSTRAINT chk_priority CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);

-- =============================================================================
-- APPROVAL ACTIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES approval_instances(id) ON DELETE CASCADE,
    step_id VARCHAR(100) NOT NULL,
    approver_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    comments TEXT,
    acted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    delegated_from UUID, -- If this action was taken by a delegate
    attachment_urls TEXT[] DEFAULT ARRAY[]::TEXT[],
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_action CHECK (action IN ('APPROVE', 'REJECT', 'DELEGATE', 'REQUEST_INFO', 'WITHDRAW', 'REASSIGN'))
);

-- =============================================================================
-- APPROVAL DELEGATES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_delegates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delegator_id UUID NOT NULL,
    delegate_id UUID NOT NULL,
    entity_types TEXT[] NOT NULL, -- What entity types can be delegated
    start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    UNIQUE(delegator_id, delegate_id, entity_types)
);

-- =============================================================================
-- APPROVAL REPORTS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES approval_workflows(id) ON DELETE CASCADE,
    entity_type VARCHAR(100) NOT NULL,
    report_date DATE NOT NULL,

    -- Metrics
    total_requests INTEGER DEFAULT 0,
    approved_count INTEGER DEFAULT 0,
    rejected_count INTEGER DEFAULT 0,
    pending_count INTEGER DEFAULT 0,
    escalated_count INTEGER DEFAULT 0,

    -- Performance metrics
    avg_processing_time INTERVAL,
    sla_breach_count INTEGER DEFAULT 0,
    sla_compliance_rate DECIMAL(5,4) DEFAULT 0,

    -- Analysis data
    bottleneck_steps JSONB DEFAULT '[]'::jsonb,
    top_approvers JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,

    UNIQUE(workflow_id, entity_type, report_date)
);

-- =============================================================================
-- APPROVAL NOTIFICATIONS QUEUE TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS approval_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES approval_instances(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL,
    notification_type VARCHAR(50) NOT NULL, -- 'PENDING_APPROVAL', 'ESCALATION', 'APPROVED', 'REJECTED'
    channel VARCHAR(20) NOT NULL, -- 'EMAIL', 'SMS', 'IN_APP', 'PUSH'
    status VARCHAR(20) DEFAULT 'PENDING',
    subject VARCHAR(500),
    message TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}'::jsonb,

    CONSTRAINT chk_notification_status CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'CANCELLED'))
);

-- =============================================================================
-- INDEXES
-- =============================================================================

-- Approval workflows indexes
CREATE INDEX IF NOT EXISTS idx_approval_workflows_entity_type ON approval_workflows(entity_type);
CREATE INDEX IF NOT EXISTS idx_approval_workflows_is_active ON approval_workflows(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_approval_workflows_is_default ON approval_workflows(entity_type, is_default) WHERE is_default = true;

-- Approval instances indexes
CREATE INDEX IF NOT EXISTS idx_approval_instances_workflow_id ON approval_instances(workflow_id);
CREATE INDEX IF NOT EXISTS idx_approval_instances_entity ON approval_instances(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_approval_instances_requested_by ON approval_instances(requested_by);
CREATE INDEX IF NOT EXISTS idx_approval_instances_status ON approval_instances(status);
CREATE INDEX IF NOT EXISTS idx_approval_instances_priority ON approval_instances(priority);
CREATE INDEX IF NOT EXISTS idx_approval_instances_current_step ON approval_instances(current_step);
CREATE INDEX IF NOT EXISTS idx_approval_instances_created_at ON approval_instances(created_at);
CREATE INDEX IF NOT EXISTS idx_approval_instances_sla_end ON approval_instances(sla_end_time) WHERE sla_end_time IS NOT NULL AND status IN ('PENDING', 'IN_PROGRESS');

-- Approval actions indexes
CREATE INDEX IF NOT EXISTS idx_approval_actions_instance_id ON approval_actions(instance_id);
CREATE INDEX IF NOT EXISTS idx_approval_actions_approver_id ON approval_actions(approver_id);
CREATE INDEX IF NOT EXISTS idx_approval_actions_action ON approval_actions(action);
CREATE INDEX IF NOT EXISTS idx_approval_actions_acted_at ON approval_actions(acted_at);
CREATE INDEX IF NOT EXISTS idx_approval_actions_delegated_from ON approval_actions(delegated_from) WHERE delegated_from IS NOT NULL;

-- Approval delegates indexes
CREATE INDEX IF NOT EXISTS idx_approval_delegates_delegator ON approval_delegates(delegator_id);
CREATE INDEX IF NOT EXISTS idx_approval_delegates_delegate ON approval_delegates(delegate_id);
CREATE INDEX IF NOT EXISTS idx_approval_delegates_active ON approval_delegates(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_approval_delegates_date_range ON approval_delegates(start_date, end_date);

-- Approval reports indexes
CREATE INDEX IF NOT EXISTS idx_approval_reports_workflow_entity ON approval_reports(workflow_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_approval_reports_date ON approval_reports(report_date);

-- Approval notifications indexes
CREATE INDEX IF NOT EXISTS idx_approval_notifications_instance ON approval_notifications(instance_id);
CREATE INDEX IF NOT EXISTS idx_approval_notifications_recipient ON approval_notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_approval_notifications_status ON approval_notifications(status);
CREATE INDEX IF NOT EXISTS idx_approval_notifications_scheduled ON approval_notifications(scheduled_at) WHERE status = 'PENDING';

-- =============================================================================
-- TRIGGERS
-- =============================================================================

-- Update updated_at timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_approval_workflows_updated_at BEFORE UPDATE ON approval_workflows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_approval_instances_updated_at BEFORE UPDATE ON approval_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

-- Function to get pending approvals for a user
CREATE OR REPLACE FUNCTION get_pending_approvals_for_user(user_id_param UUID)
RETURNS TABLE(
    instance_id UUID,
    workflow_name VARCHAR,
    entity_type VARCHAR,
    entity_id VARCHAR,
    priority VARCHAR,
    requested_by UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    sla_end_time TIMESTAMP WITH TIME ZONE,
    current_step VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ai.id as instance_id,
        aw.name as workflow_name,
        ai.entity_type,
        ai.entity_id,
        ai.priority,
        ai.requested_by,
        ai.created_at,
        ai.sla_end_time,
        ai.current_step
    FROM approval_instances ai
    JOIN approval_workflows aw ON ai.workflow_id = aw.id
    WHERE ai.status IN ('PENDING', 'IN_PROGRESS')
    AND (
        -- Direct assignment or delegation check would be implemented here
        -- This is a simplified version
        EXISTS (
            SELECT 1 FROM jsonb_array_elements(aw.steps) as step
            WHERE step->>'id' = ai.current_step
            AND (
                step->'approvers' @> jsonb_build_array(jsonb_build_object('value', user_id_param::text))
                OR EXISTS (
                    SELECT 1 FROM approval_delegates ad
                    WHERE ad.delegator_id = user_id_param
                    AND ad.is_active = true
                    AND (ad.end_date IS NULL OR ad.end_date > NOW())
                    AND ai.entity_type = ANY(ad.entity_types)
                )
            )
        )
    )
    ORDER BY
        CASE ai.priority
            WHEN 'CRITICAL' THEN 1
            WHEN 'HIGH' THEN 2
            WHEN 'MEDIUM' THEN 3
            WHEN 'LOW' THEN 4
        END,
        ai.created_at ASC;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate approval metrics
CREATE OR REPLACE FUNCTION calculate_approval_metrics(
    workflow_id_param UUID,
    entity_type_param VARCHAR,
    start_date DATE,
    end_date DATE
)
RETURNS TABLE(
    total_requests BIGINT,
    approved_count BIGINT,
    rejected_count BIGINT,
    pending_count BIGINT,
    avg_processing_time INTERVAL,
    sla_compliance_rate DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_requests,
        COUNT(*) FILTER (WHERE status = 'APPROVED') as approved_count,
        COUNT(*) FILTER (WHERE status = 'REJECTED') as rejected_count,
        COUNT(*) FILTER (WHERE status IN ('PENDING', 'IN_PROGRESS')) as pending_count,
        AVG(completed_at - created_at) FILTER (WHERE completed_at IS NOT NULL) as avg_processing_time,
        (COUNT(*) FILTER (WHERE sla_breached = false AND completed_at IS NOT NULL)::DECIMAL /
         NULLIF(COUNT(*) FILTER (WHERE completed_at IS NOT NULL), 0)) as sla_compliance_rate
    FROM approval_instances
    WHERE workflow_id = workflow_id_param
    AND entity_type = entity_type_param
    AND created_at::date BETWEEN start_date AND end_date;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- VIEWS
-- =============================================================================

-- View for approval dashboard
CREATE OR REPLACE VIEW approval_dashboard AS
SELECT
    ai.id,
    ai.entity_type,
    ai.entity_id,
    ai.status,
    ai.priority,
    ai.created_at,
    ai.sla_end_time,
    ai.sla_breached,
    aw.name as workflow_name,
    ai.current_step,
    CASE
        WHEN ai.sla_end_time IS NOT NULL AND ai.sla_end_time < NOW() AND ai.status IN ('PENDING', 'IN_PROGRESS')
        THEN true
        ELSE false
    END as is_overdue,
    EXTRACT(EPOCH FROM (COALESCE(ai.sla_end_time, NOW()) - ai.created_at)) / 3600 as hours_elapsed
FROM approval_instances ai
JOIN approval_workflows aw ON ai.workflow_id = aw.id
WHERE ai.status IN ('PENDING', 'IN_PROGRESS', 'ESCALATED');

-- View for approval performance metrics
CREATE OR REPLACE VIEW approval_performance_metrics AS
SELECT
    aw.entity_type,
    aw.name as workflow_name,
    COUNT(ai.id) as total_instances,
    COUNT(*) FILTER (WHERE ai.status = 'APPROVED') as approved_count,
    COUNT(*) FILTER (WHERE ai.status = 'REJECTED') as rejected_count,
    COUNT(*) FILTER (WHERE ai.status IN ('PENDING', 'IN_PROGRESS')) as pending_count,
    AVG(EXTRACT(EPOCH FROM (ai.completed_at - ai.created_at)) / 3600) FILTER (WHERE ai.completed_at IS NOT NULL) as avg_hours_to_complete,
    COUNT(*) FILTER (WHERE ai.sla_breached = true) as sla_breaches,
    ROUND((COUNT(*) FILTER (WHERE ai.sla_breached = false AND ai.completed_at IS NOT NULL)::DECIMAL /
           NULLIF(COUNT(*) FILTER (WHERE ai.completed_at IS NOT NULL), 0)) * 100, 2) as sla_compliance_percentage
FROM approval_workflows aw
LEFT JOIN approval_instances ai ON aw.id = ai.workflow_id
WHERE aw.is_active = true
GROUP BY aw.id, aw.entity_type, aw.name;