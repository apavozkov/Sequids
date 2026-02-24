package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	orchestrator "github.com/example/sequids/internal/orchestrator"
	orchestratorpb "github.com/example/sequids/pkg/proto/orchestratorpb"
	"google.golang.org/grpc"
)

func main() {
	listenAddr := flag.String("listen", ":9000", "gRPC listen address for the orchestrator")
	workerAddr := flag.String("worker-addr", "", "gRPC address of the worker to control")
	sensorID := flag.String("sensor-id", "", "sensor identifier to create on the worker")
	workerID := flag.String("worker-id", "worker-1", "worker identifier for scenario storage")
	intervalMs := flag.Int64("interval-ms", 1000, "sensor publish interval in milliseconds")
	flag.Parse()

	store := orchestrator.NewStore()
	server := orchestrator.NewServer(store)

	grpcServer := grpc.NewServer()
	orchestratorpb.RegisterOrchestratorServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("orchestrator listening on %s", *listenAddr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("orchestrator gRPC server error: %v", err)
		}
	}()

	if *workerAddr != "" && *sensorID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := createSensor(ctx, store, *workerAddr, *sensorID, *workerID, *intervalMs); err != nil {
			log.Printf("failed to create sensor: %v", err)
		}
	}

	waitForShutdown(grpcServer)
}

func createSensor(ctx context.Context, store *orchestrator.Store, workerAddr, sensorID, workerID string, intervalMs int64) error {
	conn, err := grpc.DialContext(ctx, workerAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := orchestratorpb.NewWorkerServiceClient(conn)
	resp, err := client.CreateSensor(ctx, &orchestratorpb.CreateSensorRequest{
		SensorId:   sensorID,
		IntervalMs: intervalMs,
		WorkerId:   workerID,
	})
	if err != nil {
		return err
	}

	store.UpsertScenario(&orchestrator.Scenario{
		SensorID:   sensorID,
		WorkerID:   workerID,
		Interval:   time.Duration(intervalMs) * time.Millisecond,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastStatus: resp.GetStatus(),
	})

	log.Printf("create sensor response: %s", resp.GetStatus())
	return nil
}

func waitForShutdown(server *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Printf("shutdown signal received")
	server.GracefulStop()
}
