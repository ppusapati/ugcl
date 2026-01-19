package services

import (
	"context"
	"fmt"
	"time"

	"p9e.in/ugcl/packages/events/domain"
)

// slaMonitorService implements ISLAMonitorService
type slaMonitorService struct {
	subscriber       *domain.EventSubscriber
	publisher        *domain.DomainEventPublisher // optional: can be used to publish monitoring events
	stopChan         chan bool
	ticker           *time.Ticker
	monitoringStatus MonitoringStatus
	checkInterval    time.Duration
}

// NewSLAMonitorService creates a new SLA monitoring service using event sourcing
func NewSLAMonitorService(subscriber *domain.EventSubscriber) ISLAMonitorService {
	return &slaMonitorService{
		subscriber:    subscriber,
		stopChan:      make(chan bool),
		checkInterval: 30 * time.Second, // Default heartbeat every 30 seconds
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

// StartSLAMonitoring begins the SLA monitoring by subscribing to domain SLA events
func (s *slaMonitorService) StartSLAMonitoring(ctx context.Context) error {
	// Heartbeat ticker (optional) for status updates
	s.ticker = time.NewTicker(s.checkInterval)
	s.monitoringStatus = MonitoringStatus{
		Service:   "sla_monitor",
		Status:    "running",
		StartTime: time.Now(),
		LastCheck: time.Now(),
	}

	// Subscribe to SLA-related events via the shared event bus
	if s.subscriber == nil {
		return fmt.Errorf("event subscriber is not initialized")
	}

	// Subscribe to SLA events
	_, err := s.subscriber.SubscribeToSLAEvents(func(c context.Context, event *domain.DomainEvent) error {
		// Update heartbeat timestamp
		s.monitoringStatus.LastCheck = time.Now()

		// Process SLA events (warning, breach, escalation, completion)
		return s.handleSLAEvent(c, event)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to SLA events: %w", err)
	}

	// Optional: subscribe to all events to aggregate metrics
	// _, _ = s.subscriber.SubscribeToAllEvents(s.ProcessEvent)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				// heartbeat only
				s.monitoringStatus.LastCheck = time.Now()
			case <-s.stopChan:
				if s.ticker != nil {
					s.ticker.Stop()
				}
				s.monitoringStatus.Status = "stopped"
				return
			}
		}
	}()

	fmt.Printf("SLA monitoring (event-sourced) started\n")
	return nil
}

// CheckSLABreaches is a no-op in event-sourced model (retained for interface compatibility)
func (s *slaMonitorService) CheckSLABreaches(ctx context.Context) error {
	return nil
}

// ProcessEscalations is a no-op in event-sourced model (retained for interface compatibility)
func (s *slaMonitorService) ProcessEscalations(ctx context.Context) error {
	return nil
}

// StopSLAMonitoring stops the SLA monitoring process
func (s *slaMonitorService) StopSLAMonitoring() {
	if s.ticker != nil {
		s.stopChan <- true
		fmt.Println("SLA monitoring stopped")
	}
}

// handleSLAEvent processes SLA-related domain events for monitoring/metrics
func (s *slaMonitorService) handleSLAEvent(ctx context.Context, event *domain.DomainEvent) error {
	// Basic routing by event type. Here we can record metrics, raise alerts, etc.
	switch event.Type {
	case domain.EventTypeSLAWarning:
		// TODO: record warning metric, potentially notify
		fmt.Printf("MONITOR: SLA warning for aggregate=%s data=%v\n", event.AggregateID, event.Data)
	case domain.EventTypeSLABreach:
		// TODO: record breach metric, trigger alert pipeline if needed
		fmt.Printf("MONITOR: SLA breach for aggregate=%s data=%v\n", event.AggregateID, event.Data)
	case domain.EventTypeSLAEscalation:
		// TODO: record escalation metric
		fmt.Printf("MONITOR: SLA escalation for aggregate=%s data=%v\n", event.AggregateID, event.Data)
	default:
		// ignore or extend with additional SLA event types when available
	}
	return nil
}
