package deps

import (
	"p9e.in/ugcl/packages/api/v1/config"
	"p9e.in/ugcl/packages/cache"
	"p9e.in/ugcl/packages/events/consumer"
	"p9e.in/ugcl/packages/events/producer"
	"p9e.in/ugcl/packages/metrics"
	"p9e.in/ugcl/packages/p9log"
	"p9e.in/ugcl/packages/timeout"
	"p9e.in/ugcl/packages/tracing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceDeps struct {
	Cfg           *config.Server
	Pool          *pgxpool.Pool
	Tp            *timeout.TimeoutProvider
	Metrics       metrics.MetricsProvider
	Log           p9log.Logger
	Cache         *cache.CacheProvider
	KafkaProducer *producer.KafkaProducer
	KafkaConsumer *consumer.KafkaConsumer
	Tracing       *tracing.TracingProvider
}

func NewServiceDeps(
	cfg *config.Server,
	pool *pgxpool.Pool,
	tp *timeout.TimeoutProvider,
	metrics metrics.MetricsProvider,
	log p9log.Logger,
	cache *cache.CacheProvider,
	kafkaProducer *producer.KafkaProducer,
	kafkaConsumer *consumer.KafkaConsumer,
	tracing *tracing.TracingProvider,
) ServiceDeps {
	return ServiceDeps{
		Cfg:           cfg,
		Pool:          pool,
		Tp:            tp,
		Metrics:       metrics,
		Log:           log,
		Cache:         cache,
		KafkaProducer: kafkaProducer,
		KafkaConsumer: kafkaConsumer,
		Tracing:       tracing,
	}
}
