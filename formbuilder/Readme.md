Let me examine all three protobuf files to understand how the workflow system works:

Based on the three protobuf files, here's how the **workflow system** works:

## 🏗️ **Architecture Overview**

The system has **3 main components** that work together:

### 1. **FormBuilder** (Form Definition)
- Creates **dynamic forms** with fields, validation, and conditional logic
- Each form has an **embedded workflow** (line 18 in formbuilder.proto)
- Supports **role-based access** and **field dependencies**

### 2. **Workflow Engine** (State Management)
- Manages **form lifecycle** through defined states
- Handles **state transitions** based on user actions
- Supports **SLA rules** and **escalations**

### 3. **FormInstance** (Runtime Data)
- Stores **actual form submissions** with field values
- Tracks **audit logs** for every state change
- Maintains **current workflow state**

## 🔄 **Workflow Flow**

```
Form Creation → Form Submission → Workflow Execution → State Transitions → Completion
```

### **Step-by-Step Process:**

**1. Form Definition Setup**
```protobuf
FormDefinition {
  workflow: {
    initial_state: "draft"
    states: ["draft", "review", "approved", "rejected"]
  }
  sla_rules: [...]
  escalations: [...]
}
```

**2. Form Submission**
```protobuf
SubmitFormRequest {
  form_id: "form-123"
  field_values: {...}  // Base64 encoded values
  action: "submit"     // Triggers workflow transition
}
```

**3. Workflow State Management**
- **Current state** stored in `FormInstance.current_state`
- **Available actions** determined by current state
- **Role-based permissions** control who can perform actions

**4. State Transitions**
```protobuf
WorkflowTransition {
  event: "approve"           // User action
  condition: "amount < 1000" // Business rule
  next_state: "approved"     // Target state
  actions: [...]             // Side effects (emails, webhooks)
}
```

## 🎯 **Key Features**

**State Types:**
- `START` - Initial form state
- `INTERMEDIATE` - Processing states
- `APPROVAL` - Requires approval
- `END` - Final states

**SLA & Escalations:**
- **Time-based rules** for each state
- **Auto-escalation** after duration
- **Multi-level escalations** with notifications

**External Integration:**
- **BPMN engines** (Camunda, Flowable, Zeebe)
- **Webhook actions** on state changes
- **Field mapping** to external systems

## 📊 **Real-World Example**

**Purchase Request Form:**
1. **Draft** → Employee fills form
2. **Manager Review** → Manager approves/rejects
3. **Finance Review** → Finance validates budget
4. **Approved/Rejected** → Final state with notifications

Each transition creates **audit logs**, triggers **notifications**, and can execute **custom actions** like updating external systems.

This design provides a **complete business process management** system integrated with dynamic form building.