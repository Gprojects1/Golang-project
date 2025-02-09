package worker

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"os"
)

func GoW() {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.Out = os.Stdout
	log.Info("Starting...")

	bootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	groupId := getEnv("KAFKA_GROUP_ID", "workers10")
	rqTopic := getEnv("KAFKA_TOPIC_RQ", "worker.rq")
	rsTopic := getEnv("KAFKA_TOPIC_RS", "worker.rs")

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupId,
		"auto.offset.reset": "latest",
		"fetch.max.bytes":         10000000,  // 10 МБ
        "max.partition.fetch.bytes": 10485760, // 10 МБ
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}
	defer consumer.Close()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"batch.size":       10485760,   // 320 КБ
        "message.max.bytes": 10000000,  // 10 МБ
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}
	defer producer.Close()

	log.Info("Serving...")
	srv := &server{}
	srv.workKmeans(rqTopic, consumer, rsTopic, producer)
}

// getEnv получает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
