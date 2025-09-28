module p9e.in/ugcl/insighthub

go 1.25.0

require (
	connectrpc.com/connect v1.18.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
	go.uber.org/fx v1.24.0
	google.golang.org/protobuf v1.36.7
	masters v0.0.0
	p9e.in/ugcl/packages v0.0.0
)

replace masters => ../masters
replace p9e.in/ugcl/packages => ../packages