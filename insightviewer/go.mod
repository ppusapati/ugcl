module p9e.in/ugcl/insightviewer

go 1.25.0

require (
	connectrpc.com/connect v1.18.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
	go.uber.org/fx v1.24.0
	google.golang.org/protobuf v1.36.7
	github.com/robfig/cron/v3 v3.0.1
	github.com/xuri/excelize/v2 v2.8.0
	masters v0.0.0
	p9e.in/ugcl/packages v0.0.0
	p9e.in/ugcl/insighthub v0.0.0
)

replace masters => ../masters
replace p9e.in/ugcl/packages => ../packages
replace p9e.in/ugcl/insighthub => ../insighthub