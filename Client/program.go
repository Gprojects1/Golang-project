package Client
/*
import (
	"context"
	"fmt"
	"log"
	"time"

	pb "awesomeProject4/protobuf/algorithms"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GoC() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:18084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAlgorithmsServiceClient(conn)

	severalSummations(c)
}

func severalSummations(c pb.AlgorithmsServiceClient) {
	startTime := time.Now()
	numRequests := 100000
	results := make(chan *pb.ClusterList, numRequests)
	errors := make(chan error, numRequests)

	points, _ := GeneratePoints(5, 2)
	Start := &pb.Kstart{Points: points, Clasters: 2}
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := c.DoAlgorithm(context.Background(), Start)
			if err != nil {
				errors <- err
			} else {
				results <- resp
			}
		}()
	}

	// Collect results and errors
	received := 0
	for received < numRequests {
		select {
		case res := <-results:
			fmt.Println("Result:", res)
			received++
		case err := <-errors:
			log.Printf("Error during summation: %v", err)
			received++ //Still count it as received, even if it was an error.
		}
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Elapsed time for several summations: %s\n", elapsedTime)
}*/
import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	pb "awesomeProject4/protobuf/algorithms"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcClient pb.AlgorithmsServiceClient

func GoC() {
	// Установить соединение с gRPC сервером
	conn, err := grpc.Dial("localhost:18084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	grpcClient = pb.NewAlgorithmsServiceClient(conn)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/generate", generateHandler)

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Генерация кластеров</title>
	</head>
	<body>
		<h1>Генерация кластеров</h1>
		<form action="/generate" method="post">
			<label>Количество точек:</label>
			<input type="number" name="numPoints" min="1" required>
			<br>
			<label>Количество кластеров:</label>
			<input type="number" name="numClusters" min="1" required>
			<br>
			<input type="submit" value="Сгенерировать">
		</form>
	</body>
	</html>
	`
	t, _ := template.New("index").Parse(tmpl)
	t.Execute(w, nil)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		numPointsStr := r.FormValue("numPoints")
		numClustersStr := r.FormValue("numClusters")

		// Преобразование строковых значений в целые числа
		numPoints, err := strconv.Atoi(numPointsStr)
		if err != nil || numPoints <= 0 {
			http.Error(w, "Некорректное количество точек", http.StatusBadRequest)
			return
		}

		numClusters, err := strconv.Atoi(numClustersStr)
		if err != nil || numClusters <= 0 {
			http.Error(w, "Некорректное количество кластеров", http.StatusBadRequest)
			return
		}

		points, _ := GeneratePoints(numPoints, 5) //!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		start := &pb.Kstart{Points: points, Clasters: int64(numClusters)}

		resp, err := grpcClient.DoAlgorithm(context.Background(), start)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Результат: %+v\n", resp)
	}
}

func GeneratePoints(numPoints int, numClusters int) ([]*pb.Point, int) {
	points := make([]*pb.Point, numPoints)
	for i := range points {
		points[i] = &pb.Point{}
		points[i].Coordinates = make([]float32, numClusters)
		for j := range points[i].Coordinates {
			points[i].Coordinates[j] = rand.Float32()
		}
	}
	return points, rand.Intn(99) + 2
}
