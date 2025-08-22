package provider

import (
	"p9e.in/ugcl/packages/events/config"
	"p9e.in/ugcl/packages/events/consumer"
	"p9e.in/ugcl/packages/events/handler"
	"p9e.in/ugcl/packages/events/producer"

	"github.com/google/wire"
)

var KafkaProviderSet = wire.NewSet(
	config.LoadConfig,
	producer.NewKafkaProducer,
	consumer.NewKafkaConsumer,
	handler.NewKafkaHandler,
)
