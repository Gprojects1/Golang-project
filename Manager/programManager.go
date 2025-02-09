package Manager

import (
	pb "awesomeProject4/protobuf/algorithms"
	"fmt"
	"awesomeProject4/Manager/kafkaAdapter"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
)

func GoM() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := getEnv("MANAGER_PORT", "18084")
	adapter, err := kafkaAdapter.NewKafkaAdapter("localhost:9092", "worker.rq", "worker.rs", "manager_group")
	if err != nil {
		log.Fatalf("failed to create Kafka adapter: %v", err)
	}

	grpcServer := grpc.NewServer()
	algorithmsService := NewAlgorithmsService(adapter)
	pb.RegisterAlgorithmsServiceServer(grpcServer, algorithmsService)

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Communication with gRPC endpoints must be made through a gRPC client. To learn how to create a client, visit: https://go.microsoft.com/fwlink/?linkid=2086909"))
	})

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("gRPC server listening on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	httpPort := 18083
	log.Printf("HTTP server listening on port %d", httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
