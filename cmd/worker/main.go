package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
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
	m := &metrics.Registry{}
	runtime := worker.NewRuntime(logger, m, mqtt.MosquittoAdapter{Host: mqttHost, Port: mqttPort}, influxURL, influxToken, influxOrg, influxBucket)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rpc.RegisterName("WorkerService", &workerService{rt: runtime, ctx: ctx}); err != nil {
		panic(err)
	}
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		panic(err)
	}
	go rpc.Accept(ln)

	register(centralRPC, workerID, listenToAdvertise(rpcAddr))
	go heartbeatLoop(ctx, centralRPC, workerID, runtime)

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Fatal(err)
		}
	}()

	logger.Info("worker started", "rpc_addr", rpcAddr, "metrics_addr", metricsAddr)
	<-ctx.Done()
}

func register(centralRPC, workerID, addr string) {
	client, err := rpc.Dial("tcp", centralRPC)
	if err != nil {
		panic(fmt.Errorf("register worker dial: %w", err))
	}
	defer client.Close()
	var out rpct.RegisterWorkerReply
	if err := client.Call("OrchestratorService.RegisterWorker", rpct.RegisterWorkerArgs{WorkerID: workerID, Address: addr}, &out); err != nil {
		panic(fmt.Errorf("register worker call: %w", err))
	}
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
