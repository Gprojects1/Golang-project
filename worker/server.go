package worker

import (
	pb "awesomeProject4/protobuf/algorithms"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

type server struct {
}

func (*server) workKmeans(rqTopic string, consumer *kafka.Consumer, rsTopic string, producer *kafka.Producer) {
	log.Printf("Subscribing on topic %s", rqTopic)
	err := consumer.Subscribe(rqTopic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic %s: %v", rqTopic, err)
	}
	for {
		record, err := consumer.ReadMessage(20 * time.Second)
		if err != nil {
			// Проверяем, является ли ошибка таймаутом
			if kafkaError, ok := err.(*kafka.Error); ok && kafkaError.Code() == kafka.ErrTimedOut {
				continue
			}
			log.Printf("Error while consuming messages: %v", err)
			continue
		}
		var request pb.WorkerRq
		err = proto.Unmarshal(record.Value, &request)
		if err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		log.Printf("Received new message from request topic with corrid=%s", request.CorrelationId)

		rsData := kMeansWithThreadPool(request.RqData, 10, 8)

		response := &pb.WorkerRs{
			RsData:        rsData,
			CorrelationId: request.CorrelationId,
		}

		log.Println("Sending result to response topic")

		responseBytes, err := proto.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			continue
		}

		// Отправляем сообщение в ответный топик
		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &rsTopic, Partition: kafka.PartitionAny},
			Value:          responseBytes,
		}, nil)
	}
}
