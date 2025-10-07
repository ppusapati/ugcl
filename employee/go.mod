module p9e.in/ugcl/employee

go 1.23

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	google.golang.org/protobuf v1.36.2
	p9e.in/ugcl/identity/user v0.0.0
	p9e.in/ugcl/packages/database/sqlc v0.0.0
)

replace p9e.in/ugcl/identity/user => ../identity/user

replace p9e.in/ugcl/packages/database/sqlc => ../packages/database/sqlc

replace p9e.in/ugcl/packages => ../packages
