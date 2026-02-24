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

	worker "github.com/apavozkov/sequids/internal/worker"
	orchestratorpb "github.com/apavozkov/sequids/pkg/proto/orchestratorpb"
	"google.golang.org/grpc"
)

func main() {
	listenAddr := flag.String("listen", ":9100", "gRPC listen address for the worker")
	workerID := flag.String("worker-id", "worker-1", "worker identifier")
	orchestratorAddr := flag.String("orchestrator-addr", "", "gRPC address of the orchestrator to report to")
	flag.Parse()

	service := worker.NewService(*workerID, *orchestratorAddr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := service.ConnectOrchestrator(ctx); err != nil {
		log.Printf("unable to connect to orchestrator: %v", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			log.Printf("error closing orchestrator connection: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	orchestratorpb.RegisterWorkerServiceServer(grpcServer, service)

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("worker %s listening on %s", *workerID, *listenAddr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("worker gRPC server error: %v", err)
		}
	}()

	waitForShutdown(grpcServer)
}

func waitForShutdown(server *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Printf("shutdown signal received")
	server.GracefulStop()
}
