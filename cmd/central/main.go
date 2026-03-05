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
		fmt.Println("usage: central [serve|push-scenario|run]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "serve":
		serve()
	case "push-scenario":
		pushScenario()
	case "run":
		runExp()
	default:
		panic("unknown command")
	}
}

func serve() {
	var rpcAddr, metricsAddr, db, formulas, anomalies string
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc-addr", ":50051", "orchestrator rpc listen address")
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080", "metrics listen address")
	fs.StringVar(&db, "db", "./sequids.db", "sqlite path")
	fs.StringVar(&formulas, "formulas", "./configs/formulas/formulas.yaml", "formulas catalog path")
	fs.StringVar(&anomalies, "anomalies", "./configs/anomalies/anomalies.yaml", "anomalies catalog path")
	_ = fs.Parse(os.Args[2:])

	logger := logging.New()
	m := &metrics.Registry{}
	store, err := storage.NewSQLiteStore(db)
	if err != nil {
		panic(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		panic(err)
	}
	cat, err := catalog.Load(formulas, anomalies)
	if err != nil {
		panic(err)
	}

	srv := orchestrator.NewServer(logger, m, store, cat)
	if err := rpc.RegisterName("OrchestratorService", srv); err != nil {
		panic(err)
	}
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		panic(err)
	}
	go rpc.Accept(ln)

	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	go func() {
		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			log.Fatal(err)
		}
	}()
	logger.Info("central started", "rpc_addr", rpcAddr, "metrics_addr", metricsAddr)
	select {}
}

func pushScenario() {
	var rpcAddr, file, name string
	fs := flag.NewFlagSet("push-scenario", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc", "localhost:50051", "central rpc address")
	fs.StringVar(&file, "file", "", "scenario dsl file")
	fs.StringVar(&name, "name", "", "scenario name")
	_ = fs.Parse(os.Args[2:])
	b, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	var out rpct.PushScenarioReply
	if err := client.Call("OrchestratorService.PushScenario", rpct.PushScenarioArgs{Name: name, DSL: string(b)}, &out); err != nil {
		panic(err)
	}
	fmt.Printf("{\"scenario_id\":\"%s\"}\n", out.ScenarioID)
}

func runExp() {
	var rpcAddr, id string
	var seed int64
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.StringVar(&rpcAddr, "rpc", "localhost:50051", "central rpc address")
	fs.StringVar(&id, "scenario", "", "scenario id")
	fs.Int64Var(&seed, "seed", time.Now().UnixNano(), "seed")
	_ = fs.Parse(os.Args[2:])
	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	var out rpct.RunExperimentReply
	if err := client.Call("OrchestratorService.RunExperiment", rpct.RunExperimentArgs{ScenarioID: id, Seed: seed}, &out); err != nil {
		panic(err)
	}
	fmt.Printf("{\"run_id\":\"%s\"}\n", out.RunID)
}
