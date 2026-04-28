package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	orchestratorv1 "sequids/api/gen/orchestratorv1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sequidsctl [start|stop|status|logs]")
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "start":
		err = startCmd()
	case "stop":
		err = stopCmd()
	case "status":
		err = statusCmd()
	case "logs":
		err = logsCmd()
	default:
		err = fmt.Errorf("unknown command")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func dial(addr string) (orchestratorv1.OrchestratorServiceClient, func() error, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return orchestratorv1.NewOrchestratorServiceClient(conn), conn.Close, nil
}

func startCmd() error {
	var grpcAddr, scenarioFile, scenarioName string
	var seed int64
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc")
	fs.StringVar(&scenarioFile, "scenario-file", "", "path to dsl scenario")
	fs.StringVar(&scenarioName, "scenario-name", "experiment", "scenario name")
	fs.Int64Var(&seed, "seed", time.Now().UnixNano(), "random seed")
	_ = fs.Parse(os.Args[2:])
	if scenarioFile == "" {
		return fmt.Errorf("-scenario-file is required")
	}
	body, err := os.ReadFile(scenarioFile)
	if err != nil {
		return err
	}
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	ps, err := client.PushScenario(context.Background(), &orchestratorv1.PushScenarioRequest{Name: scenarioName, Dsl: string(body)})
	if err != nil {
		return err
	}
	run, err := client.RunExperiment(context.Background(), &orchestratorv1.RunExperimentRequest{ScenarioId: ps.ScenarioId, Seed: seed})
	if err != nil {
		return err
	}
	fmt.Printf("started scenario_id=%s run_id=%s\n", ps.ScenarioId, run.RunId)
	return nil
}

func stopCmd() error {
	var grpcAddr, runID string
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc")
	fs.StringVar(&runID, "run-id", "", "run id")
	_ = fs.Parse(os.Args[2:])
	if runID == "" {
		return fmt.Errorf("-run-id is required")
	}
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	res, err := client.StopExperiment(context.Background(), &orchestratorv1.StopExperimentRequest{RunId: runID})
	if err != nil {
		return err
	}
	fmt.Printf("stopped=%v\n", res.Stopped)
	return nil
}

func statusCmd() error {
	var grpcAddr string
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc")
	_ = fs.Parse(os.Args[2:])
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	st, err := client.GetSystemStatus(context.Background(), &orchestratorv1.GetSystemStatusRequest{})
	if err != nil {
		return err
	}
	fmt.Printf("workers=%d runs=%d\n", len(st.Workers), len(st.Runs))
	for _, w := range st.Workers {
		fmt.Printf("worker=%s master=%v load=%d active_runs=%d active_devices=%d\n", w.WorkerId, w.IsMaster, w.VirtualLoad, w.ActiveRuns, w.ActiveDevices)
	}
	for _, r := range st.Runs {
		fmt.Printf("run=%s scenario=%s active=%v\n", r.RunId, r.ScenarioId, r.Active)
	}
	return nil
}

func logsCmd() error {
	var grpcAddr string
	var limit int
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	fs.StringVar(&grpcAddr, "grpc", "localhost:50051", "central grpc")
	fs.IntVar(&limit, "limit", 100, "number of recent lines")
	_ = fs.Parse(os.Args[2:])
	client, closeFn, err := dial(grpcAddr)
	if err != nil {
		return err
	}
	defer closeFn()
	res, err := client.GetLogs(context.Background(), &orchestratorv1.GetLogsRequest{Limit: int32(limit)})
	if err != nil {
		return err
	}
	for _, e := range res.Entries {
		fmt.Printf("%s [%s] %s\n", time.Unix(e.TsUnix, 0).UTC().Format(time.RFC3339), e.Level, e.Message)
	}
	return nil
}
