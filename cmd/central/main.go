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
	"path/filepath"
	"strings"
	"time"

	"sequids/internal/catalog"
	"sequids/internal/logging"
	"sequids/internal/metrics"
	"sequids/internal/orchestrator"
	"sequids/internal/storage"
	rpct "sequids/internal/transport/rpc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: central [serve|push-scenario|run]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		if err := serve(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "push-scenario":
		if err := pushScenario(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "run":
		if err := runExp(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(1)
	}
}

func serve() error {
	var rpcAddr, metricsAddr, db, formulas, anomalies string
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc-addr", ":50051", "orchestrator rpc listen address")
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
	if err := rpc.RegisterName("OrchestratorService", srv); err != nil {
		return fmt.Errorf("register rpc service: %w", err)
	}
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		if isAddrInUse(err) {
			return fmt.Errorf("rpc address %s already in use (stop previous central or set -rpc-addr)", rpcAddr)
		}
		return fmt.Errorf("listen rpc %s: %w", rpcAddr, err)
	}
	go rpc.Accept(ln)

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Printf("metrics server stopped: %v", err)
		}
	}()
	logger.Info("central started", "rpc_addr", rpcAddr, "metrics_addr", metricsAddr, "db", absDB)
	select {}
}

func pushScenario() error {
	var rpcAddr, file, name string
	fs := flag.NewFlagSet("push-scenario", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc", "localhost:50051", "central rpc address")
	fs.StringVar(&file, "file", "", "scenario dsl file")
	fs.StringVar(&name, "name", "", "scenario name")
	_ = fs.Parse(os.Args[2:])
	if file == "" {
		return fmt.Errorf("-file is required")
	}
	if name == "" {
		return fmt.Errorf("-name is required")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read scenario file: %w", err)
	}
	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		return fmt.Errorf("connect to central %s: %w", rpcAddr, err)
	}
	defer client.Close()
	var out rpct.PushScenarioReply
	if err := client.Call("OrchestratorService.PushScenario", rpct.PushScenarioArgs{Name: name, DSL: string(b)}, &out); err != nil {
		return fmt.Errorf("push scenario via rpc: %w", err)
	}
	fmt.Printf("{\"scenario_id\":\"%s\"}\n", out.ScenarioID)
	return nil
}

func runExp() error {
	var rpcAddr, id string
	var seed int64
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc", "localhost:50051", "central rpc address")
	fs.StringVar(&id, "scenario", "", "scenario id")
	fs.Int64Var(&seed, "seed", time.Now().UnixNano(), "seed")
	_ = fs.Parse(os.Args[2:])
	if id == "" {
		return fmt.Errorf("-scenario is required")
	}
	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		return fmt.Errorf("connect to central %s: %w", rpcAddr, err)
	}
	defer client.Close()
	var out rpct.RunExperimentReply
	if err := client.Call("OrchestratorService.RunExperiment", rpct.RunExperimentArgs{ScenarioID: id, Seed: seed}, &out); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("scenario %s not found on central %s; ensure it was pushed to this same instance and DB", id, rpcAddr)
		}
		return fmt.Errorf("run experiment via rpc: %w", err)
	}
	fmt.Printf("{\"run_id\":\"%s\"}\n", out.RunID)
	return nil
}

func isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return strings.Contains(opErr.Err.Error(), "address already in use")
	}
	return strings.Contains(err.Error(), "address already in use")
}
