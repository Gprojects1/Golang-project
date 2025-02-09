package Manager

import (
	"awesomeProject4/Manager/kafkaAdapter"
	"awesomeProject4/protobuf/algorithms"
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
)

type AlgorithmsService struct {
	algorithms.UnimplementedAlgorithmsServiceServer
	adapter *kafkaAdapter.KafkaAdapter
}

// Создаем новый конструктор для AlgorithmsService, который принимает адаптер
func NewAlgorithmsService(adapter *kafkaAdapter.KafkaAdapter) *AlgorithmsService {
	return &AlgorithmsService{adapter: adapter}
}

func (s *AlgorithmsService) DoAlgorithm(ctx context.Context, request *algorithms.Kstart) (*algorithms.ClusterList, error) {
	log.Println("Handling new request")
	correlationID := uuid.New().String()
	log.Printf("Generated corrId: %s\n", correlationID)

	err := s.adapter.ProduceAsync(correlationID, &algorithms.WorkerRq{
		CorrelationId: correlationID,
		RqData:        request,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to produce message: %v", err)
	}

	log.Println("Awaiting for result")
	resultChan := s.adapter.Consume(correlationID)
	result := <-resultChan // блокируем до получения результата

	log.Println("Receiving result")
	return result.RsData, nil
}
