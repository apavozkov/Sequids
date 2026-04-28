package orchestrator

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	orchestratorv1 "sequids/api/gen/orchestratorv1"
	"sequids/internal/catalog"
	"sequids/internal/metrics"
	"sequids/internal/scenario"
	"sequids/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type workerState struct {
	Address      string
	LastSeen     time.Time
	VirtualLoad  int
	IsMaster     bool
	AssignedRuns int
}

type runState struct {
	ScenarioID string
	StartedAt  time.Time
	Active     bool
	Workers    []string
}

type logEntry struct {
	TS      time.Time
	Level   string
	Message string
}

type Server struct {
	orchestratorv1.UnimplementedOrchestratorServiceServer
	logger  *slog.Logger
	metrics *metrics.Registry
	store   *storage.SQLiteStore
	cat     *catalog.Catalog
	mu      sync.RWMutex
	workers map[string]*workerState
	runs    map[string]*runState
	logs    []logEntry
}

func NewServer(logger *slog.Logger, metrics *metrics.Registry, store *storage.SQLiteStore, cat *catalog.Catalog) *Server {
	return &Server{logger: logger, metrics: metrics, store: store, cat: cat, workers: map[string]*workerState{}, runs: map[string]*runState{}, logs: make([]logEntry, 0, 256)}
}

func (s *Server) RegisterWorker(ctx context.Context, req *orchestratorv1.RegisterWorkerRequest) (*orchestratorv1.RegisterWorkerResponse, error) {
	s.mu.Lock()
	s.workers[req.WorkerId] = &workerState{Address: req.Address, LastSeen: time.Now().UTC()}
	s.logLocked("INFO", fmt.Sprintf("worker registered: %s (%s)", req.WorkerId, req.Address))
	s.rebalanceMasterLocked(ctx)
	s.mu.Unlock()
	s.metrics.IncEvents()
	return &orchestratorv1.RegisterWorkerResponse{Accepted: true}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *orchestratorv1.HeartbeatRequest) (*orchestratorv1.HeartbeatResponse, error) {
	s.mu.Lock()
	if st, ok := s.workers[req.WorkerId]; ok {
		st.LastSeen = time.Now().UTC()
		st.VirtualLoad = int(req.VirtualLoad)
	}
	s.rebalanceMasterLocked(ctx)
	s.mu.Unlock()
	return &orchestratorv1.HeartbeatResponse{Ok: true}, nil
}

func (s *Server) PushScenario(_ context.Context, req *orchestratorv1.PushScenarioRequest) (*orchestratorv1.PushScenarioResponse, error) {
	if _, err := scenario.ParseYAMLLike(bytes.NewBufferString(req.Dsl)); err != nil {
		return nil, fmt.Errorf("dsl parse failed: %w", err)
	}
	id := newID("scn")
	if err := s.store.SaveScenario(context.Background(), id, req.Name, req.Dsl); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.logLocked("INFO", fmt.Sprintf("scenario pushed: %s (%s)", req.Name, id))
	s.mu.Unlock()
	s.metrics.IncEvents()
	return &orchestratorv1.PushScenarioResponse{ScenarioId: id}, nil
}

func (s *Server) RunExperiment(_ context.Context, req *orchestratorv1.RunExperimentRequest) (*orchestratorv1.RunExperimentResponse, error) {
	dsl, err := s.store.GetScenario(context.Background(), req.ScenarioId)
	if err != nil {
		return nil, err
	}
	scn, err := scenario.ParseYAMLLike(bytes.NewBufferString(dsl))
	if err != nil {
		return nil, err
	}
	scn, err = s.cat.ResolveScenario(scn)
	if err != nil {
		return nil, err
	}
	applyFlowControls(&scn)
	runID := newID("run")
	workers, err := s.dispatchToWorkers(context.Background(), runID, req.Seed, scn.Devices)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.runs[runID] = &runState{ScenarioID: req.ScenarioId, StartedAt: time.Now().UTC(), Active: true, Workers: workers}
	s.logLocked("INFO", fmt.Sprintf("run started: %s scenario=%s workers=%d", runID, req.ScenarioId, len(workers)))
	s.mu.Unlock()
	s.metrics.IncEvents()
	return &orchestratorv1.RunExperimentResponse{RunId: runID}, nil
}

func applyFlowControls(scn *scenario.Scenario) {
	if scn == nil {
		return
	}
	byID := map[string]*scenario.Device{}
	for i := range scn.Devices {
		byID[scn.Devices[i].ID] = &scn.Devices[i]
	}
	type ctrl struct {
		onGT   float64
		offLT  float64
		hasOn  bool
		hasOff bool
	}
	ctrlByDevice := map[string]ctrl{}
	for _, f := range scn.Flows {
		d := byID[f.Device]
		if d == nil || d.From == "" || len(f.Conditions) == 0 {
			continue
		}
		cx := ctrlByDevice[f.Device]
		for _, c := range f.Conditions {
			if c.Threshold == nil {
				continue
			}
			for _, a := range f.Actions {
				if a.Command == "power_on" && c.Op == "gt" {
					cx.onGT = *c.Threshold
					cx.hasOn = true
				}
				if a.Command == "power_off" && c.Op == "lt" {
					cx.offLT = *c.Threshold
					cx.hasOff = true
				}
			}
		}
		ctrlByDevice[f.Device] = cx
	}
	for id, cx := range ctrlByDevice {
		d := byID[id]
		if d == nil || (!cx.hasOn && !cx.hasOff) {
			continue
		}
		d.Formula = fmt.Sprintf("control:from=%s;on_gt=%g;off_lt=%g", d.From, cx.onGT, cx.offLT)
		if d.FrequencyHz <= 0 {
			d.FrequencyHz = 1
		}
	}
}

func (s *Server) StopExperiment(_ context.Context, req *orchestratorv1.StopExperimentRequest) (*orchestratorv1.StopExperimentResponse, error) {
	s.mu.Lock()
	r, ok := s.runs[req.RunId]
	if !ok {
		s.mu.Unlock()
		return &orchestratorv1.StopExperimentResponse{Stopped: false}, nil
	}
	r.Active = false
	workers := append([]string(nil), r.Workers...)
	s.logLocked("INFO", fmt.Sprintf("run stop requested: %s", req.RunId))
	s.mu.Unlock()

	for _, wid := range workers {
		s.mu.RLock()
		ws := s.workers[wid]
		s.mu.RUnlock()
		if ws == nil {
			continue
		}
		conn, err := grpc.NewClient(ws.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}
		client := orchestratorv1.NewWorkerControlServiceClient(conn)
		_, _ = client.StopWorkload(context.Background(), &orchestratorv1.StopWorkloadRequest{RunId: req.RunId})
		_ = conn.Close()
	}
	return &orchestratorv1.StopExperimentResponse{Stopped: true}, nil
}

func (s *Server) GetSystemStatus(_ context.Context, _ *orchestratorv1.GetSystemStatusRequest) (*orchestratorv1.GetSystemStatusResponse, error) {
	s.mu.RLock()
	workers := make([]*orchestratorv1.WorkerStatus, 0, len(s.workers))
	for id, st := range s.workers {
		workers = append(workers, &orchestratorv1.WorkerStatus{WorkerId: id, Address: st.Address, IsMaster: st.IsMaster, VirtualLoad: int32(st.VirtualLoad), AssignedRuns: int32(st.AssignedRuns), LastSeenUnix: st.LastSeen.Unix()})
	}
	sort.Slice(workers, func(i, j int) bool { return workers[i].WorkerId < workers[j].WorkerId })
	runs := make([]*orchestratorv1.RunStatus, 0, len(s.runs))
	for id, st := range s.runs {
		runs = append(runs, &orchestratorv1.RunStatus{RunId: id, ScenarioId: st.ScenarioID, StartedAtUnix: st.StartedAt.Unix(), Active: st.Active})
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].RunId < runs[j].RunId })
	s.mu.RUnlock()

	for _, w := range workers {
		conn, err := grpc.NewClient(w.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}
		client := orchestratorv1.NewWorkerControlServiceClient(conn)
		res, err := client.GetWorkerStatus(context.Background(), &orchestratorv1.GetWorkerStatusRequest{})
		_ = conn.Close()
		if err != nil {
			continue
		}
		w.ActiveRuns = res.ActiveRuns
		w.ActiveDevices = res.ActiveDevices
	}
	return &orchestratorv1.GetSystemStatusResponse{Workers: workers, Runs: runs}, nil
}

func (s *Server) GetLogs(_ context.Context, req *orchestratorv1.GetLogsRequest) (*orchestratorv1.GetLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	s.mu.RLock()
	start := 0
	if len(s.logs) > limit {
		start = len(s.logs) - limit
	}
	entries := make([]*orchestratorv1.LogEntry, 0, len(s.logs)-start)
	for _, e := range s.logs[start:] {
		entries = append(entries, &orchestratorv1.LogEntry{TsUnix: e.TS.Unix(), Level: e.Level, Message: e.Message})
	}
	s.mu.RUnlock()
	return &orchestratorv1.GetLogsResponse{Entries: entries}, nil
}

func (s *Server) dispatchToWorkers(ctx context.Context, runID string, seed int64, devices []scenario.Device) ([]string, error) {
	s.mu.Lock()
	s.rebalanceMasterLocked(ctx)
	active := s.activeWorkersLocked()
	s.mu.Unlock()
	if len(active) == 0 {
		return nil, fmt.Errorf("no active workers registered")
	}
	buckets := map[string][]scenario.Device{}
	for i, d := range devices {
		wid := active[i%len(active)]
		buckets[wid] = append(buckets[wid], d)
	}
	for _, wid := range active {
		s.mu.Lock()
		ws := s.workers[wid]
		ws.AssignedRuns++
		s.mu.Unlock()

		conn, err := grpc.NewClient(ws.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("dial worker %s: %w", wid, err)
		}
		client := orchestratorv1.NewWorkerControlServiceClient(conn)
		args := &orchestratorv1.StartWorkloadRequest{RunId: runID, Seed: seed, IsMaster: ws.IsMaster}
		for _, d := range buckets[wid] {
			cfg := &orchestratorv1.DeviceConfig{Id: d.ID, Type: d.Type, Topic: d.Topic, PublishFrequencyHz: d.FrequencyHz, Formula: d.Formula, Gain: d.Gain, Offset: d.Offset, StartupDelaySec: d.StartupDelaySec, JitterRatio: d.JitterRatio}
			if d.ClampMin != nil {
				cfg.ClampMin = *d.ClampMin
			}
			if d.ClampMax != nil {
				cfg.ClampMax = *d.ClampMax
			}
			for _, a := range d.Anomalies {
				cfg.Anomalies = append(cfg.Anomalies, &orchestratorv1.AnomalyConfig{Kind: a.Kind, Probability: a.Probability, Amplitude: a.Amplitude, DriftPerSec: a.DriftPerSec, DurationSec: a.DurationSec, HoldSec: a.HoldSec})
			}
			args.Devices = append(args.Devices, cfg)
		}
		_, err = client.StartWorkload(context.Background(), args)
		_ = conn.Close()
		if err != nil {
			return nil, fmt.Errorf("start workload on %s: %w", wid, err)
		}
	}
	return active, nil
}

func (s *Server) rebalanceMasterLocked(ctx context.Context) {
	active := s.activeWorkersLocked()
	if len(active) == 0 {
		return
	}
	sort.Slice(active, func(i, j int) bool {
		wi, wj := s.workers[active[i]], s.workers[active[j]]
		if wi.VirtualLoad == wj.VirtualLoad {
			return wi.AssignedRuns < wj.AssignedRuns
		}
		return wi.VirtualLoad < wj.VirtualLoad
	})
	masterID := active[0]
	for id, st := range s.workers {
		newMaster := id == masterID
		if st.IsMaster != newMaster {
			st.IsMaster = newMaster
			go s.assignRole(ctx, id, st.Address, newMaster)
		}
	}
}

func (s *Server) assignRole(_ context.Context, workerID, addr string, isMaster bool) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("assign role dial failed", "worker", workerID, "err", err)
		return
	}
	defer conn.Close()
	client := orchestratorv1.NewWorkerControlServiceClient(conn)
	if _, err := client.AssignRole(context.Background(), &orchestratorv1.AssignRoleRequest{IsMaster: isMaster}); err != nil {
		s.logger.Error("assign role failed", "worker", workerID, "err", err)
	}
}

func (s *Server) activeWorkersLocked() []string {
	now := time.Now().UTC()
	ids := make([]string, 0, len(s.workers))
	for id, st := range s.workers {
		if now.Sub(st.LastSeen) <= 20*time.Second {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func (s *Server) logLocked(level, msg string) {
	s.logs = append(s.logs, logEntry{TS: time.Now().UTC(), Level: level, Message: msg})
	if len(s.logs) > 2000 {
		s.logs = s.logs[len(s.logs)-2000:]
	}
}

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
