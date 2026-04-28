package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	orchestratorv1 "sequids/api/gen/orchestratorv1"
	"sequids/internal/catalog"
	"sequids/internal/logging"
	"sequids/internal/metrics"
	"sequids/internal/orchestrator"
	"sequids/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: central [serve|push-scenario|run|stop|status|logs]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		errExit(serve())
	case "push-scenario":
		errExit(pushScenario())
	case "run":
		errExit(runExp())
	case "stop":
		errExit(stopExp())
	case "status":
		errExit(status())
	case "logs":
		errExit(logs())
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(1)
	}
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func serve() error {
	var grpcAddr, metricsAddr, db, formulas, anomalies string
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc-addr", ":50051", "orchestrator grpc listen address")
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080", "metrics listen address")
	fs.StringVar(&db, "db", "./sequids.db", "sqlite path")
	fs.StringVar(&formulas, "formulas", "./configs/formulas/formulas.yaml", "formulas catalog path")
	fs.StringVar(&anomalies, "anomalies", "./configs/anomalies/anomalies.yaml", "anomalies catalog path")
	_ = fs.Parse(os.Args[2:])

	absDB, _ := filepath.Abs(db)
	logger := logging.New()
	m := &metrics.Registry{}
	store, err := storage.NewSQLiteStore(db)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}
	cat, err := catalog.Load(formulas, anomalies)
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	srv := orchestrator.NewServer(logger, m, store, cat)
	g := grpc.NewServer()
	orchestratorv1.RegisterOrchestratorServiceServer(g, srv)
	ln, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		if isAddrInUse(err) {
			return fmt.Errorf("grpc address %s already in use", grpcAddr)
		}
		return fmt.Errorf("listen grpc %s: %w", grpcAddr, err)
	}
	go func() {
		if err := g.Serve(ln); err != nil {
			log.Printf("grpc server stopped: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Printf("metrics server stopped: %v", err)
		}
	}()
	logger.Info("central started", "grpc_addr", grpcAddr, "metrics_addr", metricsAddr, "db", absDB)
	select {}
}

func dial(addr string) (orchestratorv1.OrchestratorServiceClient, func() error, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return orchestratorv1.NewOrchestratorServiceClient(conn), conn.Close, nil
}

func pushScenario() error {
	var grpcAddr, file, name string
	fs := flag.NewFlagSet("push-scenario", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc address")
	fs.StringVar(&file, "file", "", "scenario dsl file")
	fs.StringVar(&name, "name", "", "scenario name")
	_ = fs.Parse(os.Args[2:])
	if file == "" || name == "" {
		return fmt.Errorf("-file and -name are required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	out, err := client.PushScenario(context.Background(), &orchestratorv1.PushScenarioRequest{Name: name, Dsl: string(b)})
	if err != nil {
		return err
	}
	fmt.Printf("{\"scenario_id\":\"%s\"}\n", out.ScenarioId)
	return nil
}

func runExp() error {
	var grpcAddr, id string
	var seed int64
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc address")
	fs.StringVar(&id, "scenario", "", "scenario id")
	fs.Int64Var(&seed, "seed", time.Now().UnixNano(), "seed")
	_ = fs.Parse(os.Args[2:])
	if id == "" {
		return fmt.Errorf("-scenario is required")
	}
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	out, err := client.RunExperiment(context.Background(), &orchestratorv1.RunExperimentRequest{ScenarioId: id, Seed: seed})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("scenario %s not found", id)
		}
		return err
	}
	fmt.Printf("{\"run_id\":\"%s\"}\n", out.RunId)
	return nil
}

func stopExp() error {
	var grpcAddr, runID string
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc address")
	fs.StringVar(&runID, "run", "", "run id")
	_ = fs.Parse(os.Args[2:])
	if runID == "" {
		return fmt.Errorf("-run is required")
	}
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	out, err := client.StopExperiment(context.Background(), &orchestratorv1.StopExperimentRequest{RunId: runID})
	if err != nil {
		return err
	}
	fmt.Printf("{\"stopped\":%v}\n", out.Stopped)
	return nil
}

func status() error {
	var grpcAddr string
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc address")
	_ = fs.Parse(os.Args[2:])
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	out, err := client.GetSystemStatus(context.Background(), &orchestratorv1.GetSystemStatusRequest{})
	if err != nil {
		return err
	}
	fmt.Printf("workers=%d runs=%d\n", len(out.Workers), len(out.Runs))
	for _, w := range out.Workers {
		fmt.Printf("worker=%s master=%v load=%d active_runs=%d active_devices=%d\n", w.WorkerId, w.IsMaster, w.VirtualLoad, w.ActiveRuns, w.ActiveDevices)
	}
	for _, r := range out.Runs {
		fmt.Printf("run=%s scenario=%s active=%v started_at=%s\n", r.RunId, r.ScenarioId, r.Active, time.Unix(r.StartedAtUnix, 0).UTC().Format(time.RFC3339))
	}
	return nil
}

func logs() error {
	var grpcAddr string
	var limit int
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc address")
	fs.IntVar(&limit, "limit", 50, "max lines")
	_ = fs.Parse(os.Args[2:])
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	out, err := client.GetLogs(context.Background(), &orchestratorv1.GetLogsRequest{Limit: int32(limit)})
	if err != nil {
		return err
	}
	for _, e := range out.Entries {
		fmt.Printf("%s [%s] %s\n", time.Unix(e.TsUnix, 0).UTC().Format(time.RFC3339), e.Level, e.Message)
	}
	return nil
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return strings.Contains(opErr.Err.Error(), "address already in use")
	}
	return strings.Contains(err.Error(), "address already in use")
}
