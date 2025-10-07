module p9e.in/ugcl/identity/entity

go 1.25.0

require (
	connectrpc.com/connect v1.18.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	go.uber.org/fx v1.24.0
	google.golang.org/protobuf v1.36.7
	p9e.in/ugcl/packages v0.0.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace (
	p9e.in/ugcl/core => ../../core
	p9e.in/ugcl/packages => ../../packages
)
