# Complete SLA Enforcement Testing Guide

## 🎯 Overview
This guide demonstrates the complete SLA enforcement system including rule creation, automatic tracking, breach detection, and escalations.

## 📋 Prerequisites
- Application running with SLA monitor service active
- Database with SLA tables created
- Bearer token for authentication

## 🔄 Complete Test Flow

### Step 1: Create SLA Rule
```bash
POST {{baseUrl}}/workflow/sla-rules
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "state": "hr_review",
  "name": "HR Review SLA",
  "description": "HR must review leave requests within 2 hours",
  "duration": {
    "value": 2,
    "unit": "HOURS"
  },
  "escalation_levels": [
    {
      "level": 1,
      "after": {
        "value": 30,
        "unit": "MINUTES"
      },
      "actions": [
        {
          "type": "email",
          "recipients": ["hr@company.com", "manager@company.com"],
          "template": "sla_breach_notification",
          "params": {
            "severity": "warning",
            "message": "Leave request requires attention"
          }
        }
      ]
    },
    {
      "level": 2,
      "after": {
        "value": 1,
        "unit": "HOURS"
      },
      "actions": [
        {
          "type": "sms",
          "recipients": ["+1234567890"],
          "template": "urgent_sla_breach",
          "params": {
            "severity": "critical"
          }
        },
        {
          "type": "reassign",
          "recipients": ["senior_hr@company.com"]
        }
      ]
    }
  ],
  "applicable_roles": ["hr_manager", "hr_admin"],
  "conditions": "priority == 'high' || department == 'executive'",
  "active": true
}
```

**Expected Response:**
```json
{
  "sla_rule": {
    "id": "uuid-generated",
    "state": "hr_review",
    "name": "HR Review SLA",
    "active": true,
    "created_at": "timestamp"
  }
}
```

### Step 2: Trigger Workflow Transition
```bash
POST {{baseUrl}}/workflow/transition
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "instance_id": "your-form-instance-id",
  "next_state": "hr_review",
  "context": {
    "comments": "Moving to HR review for approval",
    "priority": "high",
    "department": "executive"
  }
}
```

**What Happens:**
1. Form instance moves to `hr_review` state
2. System finds matching SLA rule
3. SLA instance created with due time = now + 2 hours
4. Background monitor starts tracking

### Step 3: Verify SLA Instance Creation
```bash
GET {{baseUrl}}/workflow/sla-rules/hr_review
Authorization: Bearer {{token}}
```

**Monitor Logs:**
```
SLA tracking created: Instance=uuid, Rule=uuid, Due=2024-01-15T10:30:00Z
SLA monitoring started - checking every 30 seconds
```

### Step 4: Simulate SLA Breach (Wait or Modify Time)

**Option A: Wait for actual breach**
- Wait 2+ hours for natural breach

**Option B: Simulate by modifying due time in database**
```sql
UPDATE sla_instances 
SET due_time = NOW() - INTERVAL '1 hour'
WHERE instance_id = 'your-instance-id';
```

### Step 5: Observe Breach Detection
**Monitor Logs (after 30 seconds):**
```
SLA BREACH: Instance uuid exceeded due time 2024-01-15T08:30:00Z
Escalation level 1 scheduled for 2024-01-15T09:00:00Z
Escalation level 2 scheduled for 2024-01-15T09:30:00Z
```

### Step 6: Watch Escalation Execution
**Level 1 Escalation (30 minutes after breach):**
```
EMAIL: Sending to [hr@company.com, manager@company.com] using template sla_breach_notification with params map[message:Leave request requires attention severity:warning]
Escalation level 1 completed for SLA instance uuid
```

**Level 2 Escalation (1 hour after breach):**
```
SMS: Sending to [+1234567890] using template urgent_sla_breach with params map[severity:critical]
REASSIGN: Task uuid reassigned to [senior_hr@company.com]
Escalation level 2 completed for SLA instance uuid
```

## 🔍 Verification Queries

### Check SLA Instances
```sql
SELECT 
    si.id,
    si.instance_id,
    si.state,
    si.status,
    si.start_time,
    si.due_time,
    si.breach_time,
    sr.name as sla_rule_name
FROM sla_instances si
JOIN sla_rules sr ON si.sla_rule_id = sr.id
WHERE si.instance_id = 'your-instance-id';
```

### Check Escalations
```sql
SELECT 
    sei.id,
    sei.escalation_level,
    sei.triggered_at,
    sei.status,
    sei.actions_executed
FROM sla_escalation_instances sei
JOIN sla_instances si ON sei.sla_instance_id = si.id
WHERE si.instance_id = 'your-instance-id'
ORDER BY sei.escalation_level;
```

### Check Active Monitoring
```sql
SELECT COUNT(*) as active_slas FROM sla_instances WHERE status = 'active';
SELECT COUNT(*) as breached_slas FROM sla_instances WHERE status = 'breached';
```

## 🎯 Test Scenarios

### Scenario 1: Normal Completion (No Breach)
1. Create SLA rule with 1 hour duration
2. Transition to tracked state
3. Complete workflow within 1 hour
4. Verify SLA marked as completed

### Scenario 2: Multiple SLA Rules
1. Create multiple SLA rules for same state
2. Transition to state
3. Verify multiple SLA instances created
4. Test different breach conditions

### Scenario 3: Conditional SLA Rules
1. Create SLA with conditions
2. Test transitions that match/don't match conditions
3. Verify selective SLA activation

### Scenario 4: Complex Escalations
1. Create SLA with 3+ escalation levels
2. Let it breach completely
3. Verify all escalations execute in sequence

## 🚨 Expected Behaviors

### ✅ Success Indicators
- SLA instances created on state transitions
- Background monitoring runs every 30 seconds
- Breaches detected accurately
- Escalations execute at correct times
- Multiple notification types work
- Database records updated correctly

### ❌ Error Scenarios to Test
- Invalid SLA rule creation
- Missing escalation configurations
- Database connection issues
- Malformed JSON in rules
- Concurrent breach processing

## 🔧 Troubleshooting

### Common Issues
1. **SLA not created**: Check if rule matches state and conditions
2. **No breach detection**: Verify monitor service is running
3. **Escalations not firing**: Check escalation level timing
4. **Database errors**: Verify table structure and permissions

### Debug Commands
```bash
# Check service status
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/health"

# List all SLA rules
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/workflow/sla-rules"

# Check specific state rules
curl -H "Authorization: Bearer $TOKEN" "$BASE_URL/workflow/sla-rules/hr_review"
```

## 📊 Performance Monitoring

### Key Metrics to Track
- SLA creation rate
- Breach detection latency
- Escalation execution time
- Database query performance
- Memory usage of monitor service

### Monitoring Queries
```sql
-- SLA performance stats
SELECT 
    status,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (due_time - start_time))) as avg_duration_seconds
FROM sla_instances 
GROUP BY status;

-- Recent breaches
SELECT * FROM sla_instances 
WHERE status = 'breached' 
AND breach_time > NOW() - INTERVAL '24 hours'
ORDER BY breach_time DESC;
```

This comprehensive testing guide ensures your SLA enforcement system works correctly across all scenarios!
