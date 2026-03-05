package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sequids/internal/adapters/mqtt"
	"sequids/internal/logging"
	"sequids/internal/metrics"
	rpct "sequids/internal/transport/rpc"
	"sequids/internal/worker"
)

type workerService struct {
	rt  *worker.Runtime
	ctx context.Context
}

func (w *workerService) AssignRole(args rpct.AssignRoleArgs, reply *rpct.AssignRoleReply) error {
	w.rt.SetMaster(args.IsMaster)
	reply.OK = true
	return nil
}

func (w *workerService) StartWorkload(args rpct.StartWorkloadArgs, reply *rpct.StartWorkloadReply) error {
	w.rt.SetMaster(args.IsMaster)
	w.rt.Start(w.ctx, args.RunID, args.Seed, args.Devices)
	reply.Started = true
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var workerID, rpcAddr, metricsAddr, centralRPC, mqttHost string
	var mqttPort int
	var influxURL, influxToken, influxOrg, influxBucket string
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	fs.StringVar(&workerID, "id", "worker-1", "worker id")
	fs.StringVar(&rpcAddr, "rpc-addr", ":50052", "worker rpc listen")
	fs.StringVar(&metricsAddr, "metrics-addr", ":8090", "metrics listen")
	fs.StringVar(&centralRPC, "central-rpc", "localhost:50051", "central rpc")
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

	if err := rpc.RegisterName("WorkerService", &workerService{rt: runtime, ctx: ctx}); err != nil {
		return fmt.Errorf("register worker rpc service: %w", err)
	}
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		if isAddrInUse(err) {
			return fmt.Errorf("worker rpc address %s already in use (stop previous worker or set -rpc-addr)", rpcAddr)
		}
		return fmt.Errorf("listen worker rpc %s: %w", rpcAddr, err)
	}
	go rpc.Accept(ln)

	if err := register(centralRPC, workerID, listenToAdvertise(rpcAddr)); err != nil {
		return err
	}
	go heartbeatLoop(ctx, centralRPC, workerID, runtime)

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Printf("worker metrics server stopped: %v", err)
		}
	}()

	logger.Info("worker started", "rpc_addr", rpcAddr, "metrics_addr", metricsAddr, "central_rpc", centralRPC)
	<-ctx.Done()
	return nil
}

func register(centralRPC, workerID, addr string) error {
	client, err := rpc.Dial("tcp", centralRPC)
	if err != nil {
		return fmt.Errorf("register worker dial central %s: %w", centralRPC, err)
	}
	defer client.Close()
	var out rpct.RegisterWorkerReply
	if err := client.Call("OrchestratorService.RegisterWorker", rpct.RegisterWorkerArgs{WorkerID: workerID, Address: addr}, &out); err != nil {
		return fmt.Errorf("register worker call failed: %w", err)
	}
	return nil
}

func heartbeatLoop(ctx context.Context, centralRPC, workerID string, runtime *worker.Runtime) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			client, err := rpc.Dial("tcp", centralRPC)
			if err != nil {
				continue
			}
			var out rpct.HeartbeatReply
			_ = client.Call("OrchestratorService.Heartbeat", rpct.HeartbeatArgs{WorkerID: workerID, VirtualLoad: runtime.VirtualLoad()}, &out)
			_ = client.Close()
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
