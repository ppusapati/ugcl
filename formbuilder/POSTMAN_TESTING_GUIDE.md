# FormBuilder API Testing Guide - Postman

## 🚀 Server Information
- **Base URL**: `http://localhost:10012`
- **Protocols**: Connect, gRPC, gRPC-Web
- **Content-Type**: `application/json`

## 🔐 Authentication Setup

### Method 1: API Key Authentication
```
Header: x-api-key
Values available:
- mobile-app-secret-key-123 (MobileApp)
- partner-portal-secret-key-456 (PartnerPortal) 
- internal-ops-secret-key-789 (InternalOps - Full Access)
```

### Method 2: JWT Token Authentication
```
Header: Authorization
Value: Bearer <jwt_token>
```

## 📋 Postman Environment Variables
Create these variables in Postman:
```
base_url = http://localhost:10012
api_key = internal-ops-secret-key-789
form_id = (will be set after form creation)
instance_id = (will be set after form submission)
workflow_id = (will be set after workflow creation)
```

---

## 🏗️ Step 1: FormBuilder Service Tests

### 1.1 Create Form Definition
**POST** `{{base_url}}/formbuilder.api.v2.form_builder.FormBuilder/CreateForm`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "form_definition": {
    "metadata": {
      "form_id": "employee-leave-request",
      "title": "Employee Leave Request Form",
      "description": "Form for requesting employee leave",
      "version": "1.0",
      "created_by": "admin",
      "allowed_roles": ["employee", "manager", "hr"],
      "audit": true,
      "table_name": "leave_requests",
      "module": "hr",
      "schema_version": 1,
      "core_fields": ["employee_id", "leave_type", "start_date", "end_date"]
    },
    "steps": [
      {
        "id": "basic_info",
        "label": "Basic Information",
        "order": 1,
        "fields": [
          {
            "id": "employee_id",
            "type": "TEXT",
            "label": "Employee ID",
            "required": true,
            "is_core_field": true,
            "validation": {
              "min_length": 3,
              "max_length": 10
            }
          },
          {
            "id": "leave_type",
            "type": "DROPDOWN",
            "label": "Leave Type",
            "required": true,
            "is_core_field": true,
            "validation": {
              "allowed_values": ["annual", "sick", "personal", "maternity"]
            }
          },
          {
            "id": "start_date",
            "type": "DATE",
            "label": "Start Date",
            "required": true,
            "is_core_field": true
          },
          {
            "id": "end_date",
            "type": "DATE",
            "label": "End Date",
            "required": true,
            "is_core_field": true
          },
          {
            "id": "reason",
            "type": "TEXTAREA",
            "label": "Reason for Leave",
            "required": true,
            "is_core_field": false
          }
        ]
      }
    ],
    "workflow": {
      "initial_state": "draft",
      "states": [
        {
          "id": "draft",
          "label": "Draft",
          "type": "START",
          "assigned_role": "employee",
          "actions": ["save", "submit"],
          "transitions": [
            {
              "event": "submit",
              "next_state": "manager_review",
              "actions": [
                {
                  "type": "email",
                  "params": {
                    "template": "manager_notification",
                    "recipient": "manager"
                  }
                }
              ]
            }
          ]
        },
        {
          "id": "manager_review",
          "label": "Manager Review",
          "type": "APPROVAL",
          "assigned_role": "manager",
          "actions": ["approve", "reject", "request_info"],
          "transitions": [
            {
              "event": "approve",
              "next_state": "hr_review",
              "actions": [
                {
                  "type": "email",
                  "params": {
                    "template": "hr_notification",
                    "recipient": "hr"
                  }
                }
              ]
            },
            {
              "event": "reject",
              "next_state": "rejected"
            }
          ]
        },
        {
          "id": "hr_review",
          "label": "HR Review",
          "type": "APPROVAL",
          "assigned_role": "hr",
          "actions": ["approve", "reject"],
          "transitions": [
            {
              "event": "approve",
              "next_state": "approved"
            },
            {
              "event": "reject",
              "next_state": "rejected"
            }
          ]
        },
        {
          "id": "approved",
          "label": "Approved",
          "type": "END"
        },
        {
          "id": "rejected",
          "label": "Rejected",
          "type": "END"
        }
      ]
    },
    "sla_rules": [
      {
        "id": "manager_review_sla",
        "name": "Manager Review SLA",
        "state": "manager_review",
        "duration": {
          "value": 2,
          "unit": "DAYS"
        },
        "escalation_levels": [
          {
            "level": 1,
            "after": {
              "value": 1,
              "unit": "DAYS"
            },
            "actions": [
              {
                "type": "email",
                "recipients": ["manager"],
                "template": "reminder_template"
              }
            ]
          }
        ],
        "active": true,
        "applicable_roles": ["manager"]
      }
    ]
  },
  "create_table": true
}
```

**Expected Response:**
```json
{
  "form_id": "employee-leave-request",
  "table_name": "leave_requests",
  "success": true,
  "message": "Form created successfully"
}
```

### 1.2 Get Form Definition
**POST** `{{base_url}}/formbuilder.api.v2.form_builder.FormBuilder/GetForm`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "form_id": "employee-leave-request"
}
```

### 1.3 List All Forms
**POST** `{{base_url}}/formbuilder.api.v2.form_builder.FormBuilder/ListForms`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "page": 1,
  "page_size": 10,
  "filter": "",
  "sort_by": "created_at"
}
```

---

## 📝 Step 2: FormInstance Service Tests

### 2.1 Submit Form (Create Instance)
**POST** `{{base_url}}/formbuilder.api.v2.form_instance.formSubmission/SubmitForm`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "form_id": "employee-leave-request",
  "field_values": {
    "employee_id": {
      "value": "RU1QMTIz"
    },
    "leave_type": {
      "value": "YW5udWFs"
    },
    "start_date": {
      "value": "MjAyNC0wMS0xNQ=="
    },
    "end_date": {
      "value": "MjAyNC0wMS0yMA=="
    },
    "reason": {
      "value": "RmFtaWx5IHZhY2F0aW9u"
    }
  },
  "action": "submit"
}
```

**Base64 Decoded Values:**
- `employee_id`: "EMP123"
- `leave_type`: "annual"  
- `start_date`: "2024-01-15"
- `end_date`: "2024-01-20"
- `reason`: "Family vacation"

**Expected Response:**
```json
{
  "instance_id": "uuid-generated-id",
  "current_state": "manager_review",
  "success": true,
  "errors": []
}
```

### 2.2 Get Form Instance
**POST** `{{base_url}}/formbuilder.api.v2.form_instance.formSubmission/GetFormInstance`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "instance_id": "{{instance_id}}"
}
```

### 2.3 Update Form Instance (State Transition)
**POST** `{{base_url}}/formbuilder.api.v2.form_instance.formSubmission/UpdateFormInstance`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "instance_id": "{{instance_id}}",
  "field_values": {
    "manager_comments": {
      "value": "QXBwcm92ZWQgZm9yIGFubnVhbCBsZWF2ZQ=="
    }
  },
  "action": "approve"
}
```

---

## 🔄 Step 3: Workflow Service Tests

### 3.1 Create Workflow
**POST** `{{base_url}}/formbuilder.api.v2.workflow.WorkflowService/CreateWorkflow`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "workflow": {
    "initial_state": "draft",
    "states": [
      {
        "id": "draft",
        "label": "Draft",
        "type": "START",
        "assigned_role": "employee",
        "actions": ["save", "submit"]
      },
      {
        "id": "approved",
        "label": "Approved", 
        "type": "END"
      }
    ],
    "metadata": {
      "name": "Simple Approval Workflow",
      "version": "1.0"
    }
  }
}
```

### 3.2 Get Workflow States
**POST** `{{base_url}}/formbuilder.api.v2.workflow.WorkflowService/GetWorkflowStates`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "form_id": "employee-leave-request"
}
```

### 3.3 Transition Workflow
**POST** `{{base_url}}/formbuilder.api.v2.workflow.WorkflowService/TransitionWorkflow`

**Headers:**
```
Content-Type: application/json
x-api-key: {{api_key}}
```

**Body:**
```json
{
  "instance_id": "{{instance_id}}",
  "event": "approve",
  "context": {
    "user_role": "manager",
    "comments": "Approved by manager"
  }
}
```

---

## 🧪 Step 4: Base64 Encoding Helper

Use this JavaScript in Postman Pre-request Script to encode values:

```javascript
// Helper function to encode strings to base64
function encodeBase64(str) {
    return btoa(str);
}

// Set encoded values
pm.environment.set("encoded_emp_id", encodeBase64("EMP123"));
pm.environment.set("encoded_leave_type", encodeBase64("annual"));
pm.environment.set("encoded_start_date", encodeBase64("2024-01-15"));
pm.environment.set("encoded_end_date", encodeBase64("2024-01-20"));
pm.environment.set("encoded_reason", encodeBase64("Family vacation"));
```

---

## 🔍 Step 5: Testing Scenarios

### Scenario 1: Complete Leave Request Flow
1. Create form definition ✅
2. Submit leave request ✅
3. Manager approves ✅
4. HR approves ✅
5. Check final state = "approved" ✅

### Scenario 2: Rejection Flow
1. Submit leave request ✅
2. Manager rejects ✅
3. Check final state = "rejected" ✅

### Scenario 3: SLA Testing
1. Submit form ✅
2. Wait for SLA duration ✅
3. Check escalation triggers ✅

---

## 📊 Expected Results Summary

| **Test** | **Expected Status** | **Expected State** |
|----------|-------------------|-------------------|
| Create Form | 200 OK | Form created |
| Submit Form | 200 OK | manager_review |
| Manager Approve | 200 OK | hr_review |
| HR Approve | 200 OK | approved |
| Get Instance | 200 OK | Full instance data |

Start with Step 1.1 and work through each test sequentially!
