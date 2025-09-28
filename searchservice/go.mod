module p9e.in/ugcl/searchservice

go 1.23

require (
	github.com/elastic/go-elasticsearch/v8 v8.10.1
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.2.1
	go.uber.org/fx v1.20.0
	go.uber.org/zap v1.26.0
	google.golang.org/protobuf v1.31.0
	p9e.in/ugcl/packages/events v0.0.0
	p9e.in/ugcl/packages/config v0.0.0
	p9e.in/ugcl/packages/database v0.0.0
)

replace p9e.in/ugcl/packages/events => ../packages/events
replace p9e.in/ugcl/packages/config => ../packages/config
replace p9e.in/ugcl/packages/database => ../packages/database