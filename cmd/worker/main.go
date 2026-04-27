package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	orchestratorv1 "sequids/api/gen/orchestratorv1"
	"sequids/internal/adapters/mqtt"
	"sequids/internal/logging"
	"sequids/internal/metrics"
	"sequids/internal/scenario"
	"sequids/internal/worker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type workerService struct {
	orchestratorv1.UnimplementedWorkerControlServiceServer
	rt  *worker.Runtime
	ctx context.Context
}

func (w *workerService) AssignRole(_ context.Context, args *orchestratorv1.AssignRoleRequest) (*orchestratorv1.AssignRoleResponse, error) {
	w.rt.SetMaster(args.IsMaster)
	return &orchestratorv1.AssignRoleResponse{Ok: true}, nil
}

func (w *workerService) StartWorkload(_ context.Context, args *orchestratorv1.StartWorkloadRequest) (*orchestratorv1.StartWorkloadResponse, error) {
	w.rt.SetMaster(args.IsMaster)
	devices := make([]scenario.Device, 0, len(args.Devices))
	for _, d := range args.Devices {
		cfg := scenario.Device{ID: d.Id, Type: d.Type, Topic: d.Topic, FrequencyHz: d.PublishFrequencyHz, Formula: d.Formula, Gain: d.Gain, Offset: d.Offset, StartupDelaySec: d.StartupDelaySec, JitterRatio: d.JitterRatio}
		if d.ClampMin != 0 {
			v := d.ClampMin
			cfg.ClampMin = &v
		}
		if d.ClampMax != 0 {
			v := d.ClampMax
			cfg.ClampMax = &v
		}
		for _, a := range d.Anomalies {
			cfg.Anomalies = append(cfg.Anomalies, scenario.Anomaly{Kind: a.Kind, Probability: a.Probability, Amplitude: a.Amplitude, DriftPerSec: a.DriftPerSec, DurationSec: a.DurationSec, HoldSec: a.HoldSec})
		}
		devices = append(devices, cfg)
	}
	w.rt.Start(w.ctx, args.RunId, args.Seed, devices)
	return &orchestratorv1.StartWorkloadResponse{Started: true}, nil
}

func (w *workerService) StopWorkload(_ context.Context, args *orchestratorv1.StopWorkloadRequest) (*orchestratorv1.StopWorkloadResponse, error) {
	return &orchestratorv1.StopWorkloadResponse{Stopped: w.rt.Stop(args.RunId)}, nil
}

func (w *workerService) GetWorkerStatus(_ context.Context, _ *orchestratorv1.GetWorkerStatusRequest) (*orchestratorv1.GetWorkerStatusResponse, error) {
	runs, devices := w.rt.Status()
	return &orchestratorv1.GetWorkerStatusResponse{VirtualLoad: int32(w.rt.VirtualLoad()), ActiveRuns: int32(runs), ActiveDevices: int32(devices)}, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var workerID, grpcAddr, metricsAddr, centralGRPC, mqttHost string
	var mqttPort int
	var influxURL, influxToken, influxOrg, influxBucket string
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	fs.StringVar(&workerID, "id", "worker-1", "worker id")
	fs.StringVar(&grpcAddr, "grpc-addr", ":50052", "worker grpc listen")
	fs.StringVar(&metricsAddr, "metrics-addr", ":8090", "metrics listen")
	fs.StringVar(&centralGRPC, "central-grpc", "localhost:50051", "central grpc")
	fs.StringVar(&mqttHost, "mqtt-host", "localhost", "mqtt host")
	fs.IntVar(&mqttPort, "mqtt-port", 1883, "mqtt port")
	fs.StringVar(&influxURL, "influx-url", "", "influx url")
	fs.StringVar(&influxToken, "influx-token", "", "influx token")
	fs.StringVar(&influxOrg, "influx-org", "sequids", "influx org")
	fs.StringVar(&influxBucket, "influx-bucket", "metrics", "influx bucket")
	_ = fs.Parse(os.Args[1:])

	logger := logging.New()
	if err := mqtt.EnsureClientAvailable(); err != nil {
		return err
	}
	m := &metrics.Registry{}
	runtime := worker.NewRuntime(logger, m, mqtt.MosquittoAdapter{Host: mqttHost, Port: mqttPort}, influxURL, influxToken, influxOrg, influxBucket)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	g := grpc.NewServer()
	orchestratorv1.RegisterWorkerControlServiceServer(g, &workerService{rt: runtime, ctx: ctx})
	ln, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		if isAddrInUse(err) {
			return fmt.Errorf("worker grpc address %s already in use", grpcAddr)
		}
		return fmt.Errorf("listen worker grpc %s: %w", grpcAddr, err)
	}
	go func() {
		if err := g.Serve(ln); err != nil {
			log.Printf("worker grpc server stopped: %v", err)
		}
	}()

	if err := registerWithRetry(ctx, logger, centralGRPC, workerID, listenToAdvertise(grpcAddr)); err != nil {
		return err
	}
	go heartbeatLoop(ctx, centralGRPC, workerID, runtime)

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Printf("worker metrics server stopped: %v", err)
		}
	}()

	logger.Info("worker started", "grpc_addr", grpcAddr, "metrics_addr", metricsAddr, "central_grpc", centralGRPC)
	<-ctx.Done()
	return nil
}

func register(centralGRPC, workerID, addr string) error {
	conn, err := grpc.NewClient(centralGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("register worker dial central %s: %w", centralGRPC, err)
	}
	defer conn.Close()
	client := orchestratorv1.NewOrchestratorServiceClient(conn)
	_, err = client.RegisterWorker(context.Background(), &orchestratorv1.RegisterWorkerRequest{WorkerId: workerID, Address: addr})
	return err
}

func registerWithRetry(ctx context.Context, logger *slog.Logger, centralGRPC, workerID, addr string) error {
	backoff := 2 * time.Second
	for {
		err := register(centralGRPC, workerID, addr)
		if err == nil {
			return nil
		}
		logger.Warn("worker register failed, retrying", "err", err, "retry_in", backoff.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 15*time.Second {
			backoff += time.Second
		}
	}
}

func heartbeatLoop(ctx context.Context, centralGRPC, workerID string, runtime *worker.Runtime) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			conn, err := grpc.NewClient(centralGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				continue
			}
			client := orchestratorv1.NewOrchestratorServiceClient(conn)
			_, _ = client.Heartbeat(context.Background(), &orchestratorv1.HeartbeatRequest{WorkerId: workerID, VirtualLoad: int32(runtime.VirtualLoad())})
			_ = conn.Close()
		}
	}
}

func listenToAdvertise(listen string) string {
	if len(listen) > 0 && listen[0] == ':' {
		return "127.0.0.1" + listen
	}
	return listen
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return strings.Contains(opErr.Err.Error(), "address already in use")
	}
	return strings.Contains(err.Error(), "address already in use")
}
