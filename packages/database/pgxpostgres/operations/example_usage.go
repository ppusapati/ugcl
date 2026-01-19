package operations

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/ugcl/packages/models"
)

// User represents a sample data model for demonstration
type User struct {
	ID       int64
	Username string
	Email    string
}

func ExampleDatabaseOperations() {
	// Create a database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "your_connection_string_here")
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	// Create a sample user data model
	userModel := &models.DataModel[User]{
		TableName:  "users",
		FieldNames: []string{"username", "email"},
		Values:     []User{{Username: "johndoe", Email: "john@example.com"}},
	}

	// Example 1: Basic Insert with default options
	insertedUser, err := ExecuteQuery[User](
		ctx,
		pool,
		userModel,
		QueryTypeInsert,
	)
	if err != nil {
		log.Printf("Insert failed: %v", err)
	}

	// Example 2: Insert with disabled metrics and tracing
	_, err = ExecuteQuery[User](
		ctx,
		pool,
		userModel,
		QueryTypeInsert,
	)

	// Example 3: Select with custom operation name and metrics
	_, err = ExecuteQuery[User](
		ctx,
		pool,
		userModel,
		QueryTypeSelect,
	)
	if err != nil {
		log.Printf("Select failed: %v", err)
	}

	// Example 4: Update with disabled retry
	_, err = ExecuteQuery[User](
		ctx,
		pool,
		userModel,
		QueryTypeUpdate,
		EnableRetry(false),
	)

	// Example 5: Transaction with full observability
	err = WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert user within transaction
		_, err := ExecuteQuery[User](
			ctx,
			tx,
			userModel,
			QueryTypeInsert,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user in transaction: %w", err)
		}

		// Update user within same transaction
		updateModel := &models.DataModel[User]{
			TableName:  "users",
			FieldNames: []string{"username"},
			Values:     []User{{Username: "updatedusername"}},
			Where:      "id = $1",
			WhereArgs:  []interface{}{insertedUser.ID},
		}
		_, err = ExecuteQuery[User](
			ctx,
			tx,
			updateModel,
			QueryTypeUpdate,
		)
		if err != nil {
			return fmt.Errorf("failed to update user in transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Printf("Transaction failed: %v", err)
	}

	// Example 6: Bulk Insert with specific configuration
	bulkUserModel := &models.DataModel[User]{
		TableName: "users",
		BulkValues: [][]User{
			{{Username: "user1", Email: "user1@example.com"}},
			{{Username: "user2", Email: "user2@example.com"}},
		},
		FieldNames: []string{"username", "email"},
	}
	_, err = ExecuteQuery(
		ctx,
		pool,
		bulkUserModel,
		QueryTypeBulkInsert,
		EnableRetry(true),
	)
	if err != nil {
		log.Printf("Bulk insert failed: %v", err)
	}
}

func main() {
	// Demonstrate various database operations
	ExampleDatabaseOperations()
}
