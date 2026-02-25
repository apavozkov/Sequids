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
	orchestratorpb "github.com/apavozkov/sequids/pkg/proto"
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
		log.Printf("не удаётся подключиться к оркестратору: %v", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			log.Printf("ошибка при закрытии соединения с оркестратором: %v", err)
		}
	}()

	grpcServer := grpc.NewServer(grpc.ForceServerCodec(orchestratorpb.NewJSONCodec()))
	orchestratorpb.RegisterWorkerServiceServer(grpcServer, service)

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("ошибка при прослушивании: %v", err)
	}

	go func() {
		log.Printf("воркер %s слушает %s", *workerID, *listenAddr)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("ошибка gRPC сервера воркера: %v", err)
		}
	}()

	waitForShutdown(grpcServer)
}

func waitForShutdown(server *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Printf("получен сигнал выключения")
	server.GracefulStop()
}
