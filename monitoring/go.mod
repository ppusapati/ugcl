module p9e.in/ugcl/monitoring

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	go.uber.org/fx v1.24.0
	p9e.in/ugcl/formbuilder v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.5.1 // indirect
	github.com/sqlc-dev/pqtype v0.3.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.71.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	p9e.in/ugcl/identity v0.0.0-00010101000000-000000000000 // indirect
	p9e.in/ugcl/packages v0.0.0-00010101000000-000000000000 // indirect
	p9e.in/ugcl/projects v0.0.0-00010101000000-000000000000 // indirect
	p9e.in/ugcl/vendors v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	p9e.in/ugcl/formbuilder => ../formbuilder
	p9e.in/ugcl/identity => ../identity
	p9e.in/ugcl/packages => ../packages
	p9e.in/ugcl/projects => ../projects
	p9e.in/ugcl/vendors => ../vendors
)
