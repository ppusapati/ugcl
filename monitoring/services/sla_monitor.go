package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "p9e.in/ugcl/formbuilder/db/generated"
	"p9e.in/ugcl/formbuilder/repository"
)

// slaMonitorService implements ISLAMonitorService
type slaMonitorService struct {
	workflowRepo     repository.IWorkflowRepository
	stopChan         chan bool
	ticker           *time.Ticker
	monitoringStatus MonitoringStatus
	checkInterval    time.Duration
}

// NewSLAMonitorService creates a new SLA monitoring service
func NewSLAMonitorService(workflowRepo repository.IWorkflowRepository) ISLAMonitorService {
	return &slaMonitorService{
		workflowRepo:  workflowRepo,
		stopChan:      make(chan bool),
		checkInterval: 30 * time.Second, // Default 30 seconds
		monitoringStatus: MonitoringStatus{
			Service: "sla_monitor",
			Status:  "stopped",
		},
	}
}

// Start implements IMonitoringService
func (s *slaMonitorService) Start(ctx context.Context) error {
	return s.StartSLAMonitoring(ctx)
}

// Stop implements IMonitoringService
func (s *slaMonitorService) Stop(ctx context.Context) error {
	s.StopSLAMonitoring()
	return nil
}

// GetStatus implements IMonitoringService
func (s *slaMonitorService) GetStatus() MonitoringStatus {
	return s.monitoringStatus
}

// SetMonitoringInterval implements ISLAMonitorService
func (s *slaMonitorService) SetMonitoringInterval(interval time.Duration) {
	s.checkInterval = interval
	if s.ticker != nil {
		s.ticker.Reset(interval)
	}
}

// StartSLAMonitoring begins the SLA monitoring process
func (s *slaMonitorService) StartSLAMonitoring(ctx context.Context) error {
	s.ticker = time.NewTicker(s.checkInterval)
	s.monitoringStatus = MonitoringStatus{
		Service:   "sla_monitor",
		Status:    "running",
		StartTime: time.Now(),
		LastCheck: time.Now(),
	}

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.monitoringStatus.LastCheck = time.Now()

				if err := s.CheckSLABreaches(ctx); err != nil {
					fmt.Printf("ERROR: SLA breach check failed: %v\n", err)
					s.monitoringStatus.Status = "error"
					s.monitoringStatus.Error = err.Error()
				} else {
					s.monitoringStatus.Status = "running"
					s.monitoringStatus.Error = ""
				}

				if err := s.ProcessEscalations(ctx); err != nil {
					fmt.Printf("ERROR: Escalation processing failed: %v\n", err)
				}

			case <-s.stopChan:
				s.ticker.Stop()
				s.monitoringStatus.Status = "stopped"
				return
			}
		}
	}()

	fmt.Printf("SLA monitoring started - checking every %v\n", s.checkInterval)
	return nil
}

// CheckSLABreaches identifies and processes SLA breaches
func (s *slaMonitorService) CheckSLABreaches(ctx context.Context) error {
	// Get all overdue SLA instances
	overdueInstances, err := s.workflowRepo.GetOverdueSLAInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to get overdue SLA instances: %w", err)
	}

	for _, instance := range overdueInstances {
		// Mark as breached if not already
		if instance.Status == "active" {
			_, err := s.workflowRepo.BreachSLAInstance(ctx, instance.ID)
			if err != nil {
				fmt.Printf("WARNING: Failed to breach SLA instance %s: %v\n", instance.ID, err)
				continue
			}

			fmt.Printf("SLA BREACH: Instance %s exceeded due time %s\n",
				instance.InstanceID, instance.DueTime.Format(time.RFC3339))

			// Trigger escalations for this breach
			if err := s.triggerEscalations(ctx, instance); err != nil {
				fmt.Printf("WARNING: Failed to trigger escalations for instance %s: %v\n", instance.ID, err)
			}
		}
	}

	return nil
}

// ProcessEscalations handles pending escalation actions
func (s *slaMonitorService) ProcessEscalations(ctx context.Context) error {
	// Get pending escalations
	pendingEscalations, err := s.workflowRepo.GetPendingEscalations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending escalations: %w", err)
	}

	for _, escalation := range pendingEscalations {
		if err := s.executeEscalationActions(ctx, escalation); err != nil {
			fmt.Printf("WARNING: Failed to execute escalation %s: %v\n", escalation.ID, err)
		}
	}

	return nil
}

// StopSLAMonitoring stops the SLA monitoring process
func (s *slaMonitorService) StopSLAMonitoring() {
	if s.ticker != nil {
		s.stopChan <- true
		fmt.Println("SLA monitoring stopped")
	}
}

// triggerEscalations creates escalation instances based on SLA rule configuration
func (s *slaMonitorService) triggerEscalations(ctx context.Context, slaInstance *db.SlaInstance) error {
	// Get the SLA rule to find escalation levels
	slaRule, err := s.workflowRepo.GetSLARule(ctx, slaInstance.SlaRuleID)
	if err != nil {
		return fmt.Errorf("failed to get SLA rule: %w", err)
	}

	// Parse escalation levels from JSON
	var escalationLevels []struct {
		Level int32 `json:"level"`
		After struct {
			Value int32  `json:"value"`
			Unit  string `json:"unit"`
		} `json:"after"`
		Actions []struct {
			Type       string            `json:"type"`
			Recipients []string          `json:"recipients"`
			Template   string            `json:"template"`
			Params     map[string]string `json:"params"`
		} `json:"actions"`
	}

	if err := json.Unmarshal(slaRule.EscalationLevels, &escalationLevels); err != nil {
		return fmt.Errorf("failed to parse escalation levels: %w", err)
	}

	now := time.Now()

	// Create escalation instances for each level
	for _, level := range escalationLevels {
		// Calculate when this escalation should trigger
		var triggerTime time.Time
		switch level.After.Unit {
		case "MINUTES":
			triggerTime = slaInstance.DueTime.Add(time.Duration(level.After.Value) * time.Minute)
		case "HOURS":
			triggerTime = slaInstance.DueTime.Add(time.Duration(level.After.Value) * time.Hour)
		case "DAYS":
			triggerTime = slaInstance.DueTime.Add(time.Duration(level.After.Value) * 24 * time.Hour)
		default:
			triggerTime = slaInstance.DueTime.Add(time.Duration(level.After.Value) * time.Hour)
		}

		// Create escalation instance
		actionsJSON, _ := json.Marshal(level.Actions)
		escalationParams := db.CreateSLAEscalationInstanceParams{
			ID:              uuid.New(),
			SlaInstanceID:   slaInstance.ID,
			EscalationLevel: level.Level,
			TriggeredAt:     triggerTime,
			ActionsExecuted: actionsJSON,
			Status:          "pending",
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		_, err := s.workflowRepo.CreateSLAEscalationInstance(ctx, escalationParams)
		if err != nil {
			fmt.Printf("WARNING: Failed to create escalation instance: %v\n", err)
			continue
		}

		fmt.Printf("Escalation level %d scheduled for %s\n", level.Level, triggerTime.Format(time.RFC3339))
	}

	return nil
}

// executeEscalationActions performs the actual escalation actions (email, SMS, etc.)
func (s *slaMonitorService) executeEscalationActions(ctx context.Context, escalation *db.SlaEscalationInstance) error {
	// Parse actions from JSON
	var actions []struct {
		Type       string            `json:"type"`
		Recipients []string          `json:"recipients"`
		Template   string            `json:"template"`
		Params     map[string]string `json:"params"`
	}

	if err := json.Unmarshal(escalation.ActionsExecuted, &actions); err != nil {
		return fmt.Errorf("failed to parse escalation actions: %w", err)
	}

	// Execute each action
	for _, action := range actions {
		switch action.Type {
		case "email":
			if err := s.sendEmailNotification(action.Recipients, action.Template, action.Params); err != nil {
				fmt.Printf("WARNING: Failed to send email notification: %v\n", err)
			}
		case "sms":
			if err := s.sendSMSNotification(action.Recipients, action.Template, action.Params); err != nil {
				fmt.Printf("WARNING: Failed to send SMS notification: %v\n", err)
			}
		case "reassign":
			if err := s.reassignTask(escalation.SlaInstanceID, action.Recipients); err != nil {
				fmt.Printf("WARNING: Failed to reassign task: %v\n", err)
			}
		case "notify":
			if err := s.sendInAppNotification(action.Recipients, action.Template, action.Params); err != nil {
				fmt.Printf("WARNING: Failed to send in-app notification: %v\n", err)
			}
		}
	}

	// Mark escalation as completed
	completeParams := db.CompleteEscalationParams{
		ID:              escalation.ID,
		ActionsExecuted: escalation.ActionsExecuted,
	}

	_, err := s.workflowRepo.CompleteEscalation(ctx, completeParams)
	if err != nil {
		return fmt.Errorf("failed to mark escalation as completed: %w", err)
	}

	fmt.Printf("Escalation level %d completed for SLA instance %s\n",
		escalation.EscalationLevel, escalation.SlaInstanceID)

	return nil
}

// Notification methods (placeholder implementations - can be extended with actual providers)
func (s *slaMonitorService) sendEmailNotification(recipients []string, template string, params map[string]string) error {
	fmt.Printf("EMAIL: Sending to %v using template %s with params %v\n", recipients, template, params)
	// TODO: Implement actual email sending (SendGrid, AWS SES, etc.)
	return nil
}

func (s *slaMonitorService) sendSMSNotification(recipients []string, template string, params map[string]string) error {
	fmt.Printf("SMS: Sending to %v using template %s with params %v\n", recipients, template, params)
	// TODO: Implement actual SMS sending (Twilio, AWS SNS, etc.)
	return nil
}

func (s *slaMonitorService) reassignTask(slaInstanceID uuid.UUID, newAssignees []string) error {
	fmt.Printf("REASSIGN: Task %s reassigned to %v\n", slaInstanceID, newAssignees)
	// TODO: Implement actual task reassignment logic
	return nil
}

func (s *slaMonitorService) sendInAppNotification(recipients []string, template string, params map[string]string) error {
	fmt.Printf("NOTIFICATION: Sending to %v using template %s with params %v\n", recipients, template, params)
	// TODO: Implement actual in-app notifications (WebSocket, push notifications, etc.)
	return nil
}
