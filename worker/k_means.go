package worker

import (
	pb "awesomeProject4/protobuf/algorithms"
	"log"
	"math/rand"
	"sync"
)

type Worker struct {
	ID       int
	JobChan  chan []*pb.Point
	wg       *sync.WaitGroup
	mu       *sync.Mutex
	clusters []*pb.Cluster
}

func euclideanDistance(p1, p2 *pb.Point) float32 {
	if len(p1.Coordinates) != len(p2.Coordinates) {
		panic("Points must have the same dimension")
	}
	var sum float32 = 0.0
	for i := range p1.Coordinates {
		diff := p1.Coordinates[i] - p2.Coordinates[i]
		sum += diff * diff
	}
	return sum
}

func initializeClusters(points []*pb.Point, k int64) []*pb.Cluster {
	// Создаем срез кластеров с длиной k
	clusters := make([]*pb.Cluster, k)

	// Инициализируем каждый элемент среза
	for i := int64(0); i < k; i++ {
		clusters[i] = &pb.Cluster{}                           // Инициализация нового объекта Cluster
		clusters[i].Centroid = points[rand.Intn(len(points))] // Установка случайного центроида
	}

	return clusters
}

func updateCentroids(clusters []*pb.Cluster) {
	for i := range clusters {
		if len(clusters[i].Points) == 0 {
			continue
		}
		dimensions := len(clusters[i].Points[0].Coordinates)
		sum := make([]float32, dimensions)
		for _, point := range clusters[i].Points {
			for d := range point.Coordinates {
				sum[d] += point.Coordinates[d]
			}
		}
		for d := range sum {
			sum[d] /= float32(len(clusters[i].Points))
		}
		clusters[i].Centroid = &pb.Point{Coordinates: sum}
	}
}

func (w *Worker) Start() {
	for points := range w.JobChan {
		assignPointsToClusters(points, w.clusters, w.wg, w.mu)
	}
}

func assignPointsToClusters(points []*pb.Point, clusters []*pb.Cluster, wg *sync.WaitGroup, mu *sync.Mutex) {
	if points == nil {
		log.Println("Ошибка: points равен nil")
	}
	if clusters == nil {
		log.Println("Ошибка: clusters равен nil")
	}
	defer wg.Done()

	tempClusters := make([][]*pb.Point, len(clusters))

	for _, point := range points {
		minDistance := euclideanDistance(point, clusters[0].Centroid)
		closestCluster := 0
		for j := 1; j < len(clusters); j++ {
			distance := euclideanDistance(point, clusters[j].Centroid)
			if distance < minDistance {
				minDistance = distance
				closestCluster = j
			}
		}
		tempClusters[closestCluster] = append(tempClusters[closestCluster], point)
	}

	mu.Lock()
	for i := range clusters {
		clusters[i].Points = append(clusters[i].Points, tempClusters[i]...)
	}
	mu.Unlock()
}

func kMeansWithThreadPool(Alfa *pb.Kstart, maxIterations int, numWorkers int) *pb.ClusterList {
	points := Alfa.Points
	k := Alfa.Clasters
	clusters := initializeClusters(points, k)

	for iteration := 0; iteration < maxIterations; iteration++ {
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := range clusters {
			clusters[i].Points = nil
		}

		// Создаем пул потоков
		jobChan := make(chan []*pb.Point, numWorkers)
		workers := make([]*Worker, numWorkers)

		for i := 0; i < numWorkers; i++ {
			workers[i] = &Worker{
				ID:       i,
				JobChan:  jobChan,
				wg:       &wg,
				mu:       &mu,
				clusters: clusters,
			}
			go workers[i].Start()
		}

		// Разделение работы между работниками
		chunkSize := (len(points) + numWorkers - 1) / numWorkers
		for i := 0; i < len(points); i += chunkSize {
			end := i + chunkSize
			if end > len(points) {
				end = len(points)
			}
			wg.Add(1)
			jobChan <- points[i:end]
		}

		close(jobChan)
		wg.Wait()

		updateCentroids(clusters)
	}
	clusterList := &pb.ClusterList{
		Clusters: clusters, // Здесь clusters должен быть типом []*pb.Cluster
	}
	return clusterList
}

func generateRandomPoints(numPoints, dimensions int) []pb.Point {
	points := make([]pb.Point, numPoints)
	for i := 0; i < numPoints; i++ {
		coords := make([]float32, dimensions)
		for d := 0; d < dimensions; d++ {
			coords[d] = rand.Float32() * 100
		}
		points[i] = pb.Point{Coordinates: coords}
	}
	return points
}
