-- Create SLA views for monitoring and reporting

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
    COUNT(CASE WHEN si.status = 'completed' AND si.completion_time <= si.due_time THEN 1 END) as on_time_completions,
    COUNT(CASE WHEN si.status = 'breached' OR (si.status = 'completed' AND si.completion_time > si.due_time) THEN 1 END) as breached_instances,
    COUNT(CASE WHEN si.status = 'active' AND NOW() > si.due_time THEN 1 END) as currently_overdue,
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
