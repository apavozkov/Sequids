package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	orchestratorpb "github.com/example/sequids/pkg/proto/orchestratorpb"
	"google.golang.org/grpc"
)

type Sensor struct {
	id       string
	interval time.Duration
	stopCh   chan struct{}
}

type Service struct {
	orchestratorpb.UnimplementedWorkerServiceServer
	workerID             string
	orchestratorAddr     string
	orchestratorClient   orchestratorpb.OrchestratorServiceClient
	orchestratorConn     *grpc.ClientConn
	sensors              map[string]*Sensor
	mu                   sync.Mutex
	rng                  *rand.Rand
}

func NewService(workerID, orchestratorAddr string) *Service {
	return &Service{
		workerID:         workerID,
		orchestratorAddr: orchestratorAddr,
		sensors:          make(map[string]*Sensor),
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Service) ConnectOrchestrator(ctx context.Context) error {
	if s.orchestratorAddr == "" {
		return nil
	}
	conn, err := grpc.DialContext(ctx, s.orchestratorAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.orchestratorConn = conn
	s.orchestratorClient = orchestratorpb.NewOrchestratorServiceClient(conn)
	return nil
}

func (s *Service) Close() error {
	if s.orchestratorConn != nil {
		return s.orchestratorConn.Close()
	}
	return nil
}

func (s *Service) CreateSensor(ctx context.Context, req *orchestratorpb.CreateSensorRequest) (*orchestratorpb.CreateSensorResponse, error) {
	interval := time.Duration(req.GetIntervalMs()) * time.Millisecond
	if interval <= 0 {
		interval = time.Second
	}

	s.mu.Lock()
	if _, exists := s.sensors[req.GetSensorId()]; exists {
		s.mu.Unlock()
		return &orchestratorpb.CreateSensorResponse{Status: "already_exists"}, nil
	}
	sensor := &Sensor{
		id:       req.GetSensorId(),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
	s.sensors[req.GetSensorId()] = sensor
	s.mu.Unlock()

	log.Printf("worker %s created sensor %s with interval %s", s.workerID, sensor.id, sensor.interval)
	go s.runSensor(sensor)

	return &orchestratorpb.CreateSensorResponse{Status: "created"}, nil
}

func (s *Service) runSensor(sensor *Sensor) {
	ticker := time.NewTicker(sensor.interval)
	defer ticker.Stop()

	s.reportStatus(sensor.id, "created", 0)

	for {
		select {
		case <-sensor.stopCh:
			return
		case <-ticker.C:
			value := s.rng.Float64() * 100
			fmt.Printf("sensor=%s value=%.4f\n", sensor.id, value)
			s.reportStatus(sensor.id, "published", value)
		}
	}
}

func (s *Service) reportStatus(sensorID, status string, value float64) {
	if s.orchestratorClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := s.orchestratorClient.ReportSensorStatus(ctx, &orchestratorpb.ReportSensorStatusRequest{
		SensorId: sensorID,
		Status:   status,
		Value:    value,
	})
	if err != nil {
		log.Printf("failed to report status to orchestrator: %v", err)
	}
}
