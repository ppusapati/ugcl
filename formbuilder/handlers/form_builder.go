package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	forminstancev2 "p9e.in/ugcl/formbuilder/api/v2/form_instance"
	workflowv2 "p9e.in/ugcl/formbuilder/api/v2/workflow"
)

// ExternalWorkflowClient handles integration with BPMN engines
type ExternalWorkflowClient struct {
	config *workflowv2.WorkflowIntegration
	client *http.Client
}

func NewExternalWorkflowClient(config *workflowv2.WorkflowIntegration) *ExternalWorkflowClient {
	return &ExternalWorkflowClient{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// TriggerWorkflow starts a workflow in external BPMN engine
func (ewc *ExternalWorkflowClient) TriggerWorkflow(
	formInstance *forminstancev2.FormInstance,
	action string,
) (string, error) {
	if !ewc.config.Enabled {
		return "", nil
	}

	// Map form fields to process variables
	variables := make(map[string]interface{})
	for formField, processVar := range ewc.config.FieldMappings {
		if value, ok := formInstance.FieldValues[formField]; ok {
			variables[processVar] = value
		}
	}

	// Add metadata
	variables["formId"] = formInstance.FormId
	variables["formInstanceId"] = formInstance.Id
	variables["action"] = action

	// Build request based on engine type
	var endpoint string
	var payload interface{}

	switch ewc.config.EngineType {
	case "camunda":
		endpoint = fmt.Sprintf("%s/process-definition/key/%s/start",
			ewc.config.Endpoint,
			ewc.config.ProcessDefinitionKey)
		payload = map[string]interface{}{
			"variables": variables,
		}
	case "flowable":
		endpoint = fmt.Sprintf("%s/runtime/process-instances", ewc.config.Endpoint)
		payload = map[string]interface{}{
			"processDefinitionKey": ewc.config.ProcessDefinitionKey,
			"variables":            variables,
		}
	case "zeebe":
		// Zeebe uses gRPC, would need different client
		return "", fmt.Errorf("Zeebe integration not implemented")
	default:
		return "", fmt.Errorf("unknown workflow engine: %s", ewc.config.EngineType)
	}

	// Make HTTP request
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ewc.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("workflow engine returned status %d", resp.StatusCode)
	}

	// Parse response to get process instance ID
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	processInstanceId, _ := result["id"].(string)
	return processInstanceId, nil
}
