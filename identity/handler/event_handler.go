package handler

import (
	"context"
	"encoding/json"
	"fmt"

	msgp "p9e.in/ugcl/packages/api/v1/message"
	"p9e.in/ugcl/packages/events/consumer"
	"p9e.in/ugcl/packages/events/producer"
	"p9e.in/ugcl/packages/p9log"
)

type UserEventHandler struct {
	log           p9log.Helper
	kafkaProducer *producer.KafkaProducer
	kafkaConsumer *consumer.KafkaConsumer
}

func NewUserEventHandler(
	log p9log.Logger,
	kafkaProducer *producer.KafkaProducer,
	kafkaConsumer *consumer.KafkaConsumer,
) *UserEventHandler {
	return &UserEventHandler{
		log:           *p9log.NewHelper(p9log.With(log, "Event Hanlder")),
		kafkaProducer: kafkaProducer,
		kafkaConsumer: kafkaConsumer,
	}
}

// Topic-specific handlers
func (e *UserEventHandler) HandleTenantConfig(ctx context.Context, msg *msgp.EventMessage) error {
	e.log.Infof("Handling tenant config message: %s", msg.Value)
	// Implement tenant config logic
	return nil
}

func (e *UserEventHandler) CreateTenant(ctx context.Context, msg *msgp.EventMessage) error {
	e.log.Infof("Creating tenant from message: %s", msg.Value)
	// return pgx_migrations.SetTenantDB(ctx, e.cfgd, "tenant002")
	return nil
}

func (e *UserEventHandler) HandleUserCreated(ctx context.Context, msg *msgp.EventMessage) error {
	// Deserialize the message
	var userData map[string]interface{}
	err := json.Unmarshal([]byte(msg.Value), &userData)
	if err != nil {
		e.log.Errorf("Failed to unmarshal user data: %v", err)
		return err
	}

	// Extract necessary user information
	email, ok := userData["email"].(string)
	if !ok {
		e.log.Errorf("Invalid email in user data: %v", userData)
		return fmt.Errorf("invalid email in user data")
	}

	username, _ := userData["username"].(string)

	// Send welcome email
	err = e.sendWelcomeEmail(ctx, email, username)
	if err != nil {
		e.log.Errorf("Failed to send welcome email to %s: %v", email, err)
		return err
	}

	e.log.Infof("Welcome email sent to user: %s", email)
	return nil
}

func (e *UserEventHandler) HandleOrderReceived(ctx context.Context, msg *msgp.EventMessage) error {
	// Deserialize the message
	var orderData map[string]interface{}
	err := json.Unmarshal([]byte(msg.Value), &orderData)
	if err != nil {
		e.log.Errorf("Failed to unmarshal order data: %v", err)
		return err
	}

	// Extract user ID and other necessary information
	userID, ok := orderData["user_id"].(string)
	if !ok {
		e.log.Errorf("Invalid user ID in order data: %v", orderData)
		return fmt.Errorf("invalid user ID in order data")
	}

	// Additional order details
	orderID, _ := orderData["order_id"].(string)
	orderTotal, _ := orderData["total"].(float64)

	fmt.Println(orderTotal)
	fmt.Println(orderID)

	e.log.Infof("Order received and processed for user: %s", userID)
	return nil
}

func (e *UserEventHandler) sendWelcomeEmail(ctx context.Context, email, username string) error {
	// Implement email sending logic
	// This is a placeholder - you'll need to integrate with your email service
	emailContent := fmt.Sprintf(`
		Dear %s,

		Welcome to YourPlanter! We're excited to have you on board.
		Your account has been successfully created.

		Best regards,
		YourPlanter Team
	`, username)

	// TODO: Replace with actual email sending mechanism
	e.log.Infof("Would send email to %s with content: %s", email, emailContent)
	defer ctx.Done()
	return nil
}
