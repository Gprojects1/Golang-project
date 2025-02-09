package kafkaAdapter

import (
	"awesomeProject4/protobuf/algorithms"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"sync"
)

type KafkaAdapter struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	results  sync.Map
	topicRq  string
	topicRs  string
}

func NewKafkaAdapter(bootstrapServers, topicRq, topicRs, groupId string) (*KafkaAdapter, error) {
	//producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"batch.size":       10485760,   // 320 КБ
        "message.max.bytes": 10000000,  // 10 МБ
	})
	if err != nil {
		return nil, err
	}
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
		"fetch.max.bytes":         10000000,  // 10 МБ
        "max.partition.fetch.bytes": 10485760, // 10 МБ
	})
	if err != nil {
		return nil, err
	}

	if err := consumer.Subscribe(topicRs, nil); err != nil {
		return nil, err
	}

	adapter := &KafkaAdapter{
		producer: producer,
		consumer: consumer,
		topicRq:  topicRq,
		topicRs:  topicRs,
	}

	go adapter.startConsuming()

	return adapter, nil
}

func (ka *KafkaAdapter) ProduceAsync(guid string, workerRq *algorithms.WorkerRq) error {
	value, err := proto.Marshal(workerRq)
	if err != nil {
		return err
	}

	// Отправка сообщения
	err = ka.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &ka.topicRq, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)

	if err != nil {
		return err
	}

	// Изменение типа канала на *algorithms.WorkerRs
	ka.results.Store(guid, make(chan *algorithms.WorkerRs, 1))

	return nil
}

func (ka *KafkaAdapter) Consume(correlationId string) <-chan *algorithms.WorkerRs {
	ch := make(chan *algorithms.WorkerRs, 1)
	ka.results.Store(correlationId, ch)
	return ch
}

func (ka *KafkaAdapter) startConsuming() {
	for {
		msg, err := ka.consumer.ReadMessage(-1)
		if err == nil {
			var workerRs algorithms.WorkerRs
			if err := proto.Unmarshal(msg.Value, &workerRs); err != nil {
				fmt.Printf("Ошибка десериализации сообщения: %v\n", err)
				continue
			}

			guid := workerRs.CorrelationId // Предполагается, что CorrelationId хранится в WorkerRs
			if resultChan, ok := ka.results.Load(guid); ok {
				// Приведение типа к *algorithms.WorkerRs
				resultChan.(chan *algorithms.WorkerRs) <- &workerRs // Отправляем указатель
				// Не закрываем канал здесь, чтобы позволить обработать несколько сообщений
			}
		} else {
			fmt.Printf("Ошибка при получении сообщения: %v\n", err)
		}
	}
}
