package orchestrator

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/rpc"
	"sort"
	"sync"
	"time"

	rpct "sequids/internal/transport/rpc"

	"sequids/internal/catalog"
	"sequids/internal/metrics"
	"sequids/internal/scenario"
	"sequids/internal/storage"
)

type workerState struct {
	Address      string
	LastSeen     time.Time
	VirtualLoad  int
	IsMaster     bool
	AssignedRuns int
}

type Server struct {
	logger  *slog.Logger
	metrics *metrics.Registry
	store   *storage.SQLiteStore
	cat     *catalog.Catalog
	mu      sync.RWMutex
	workers map[string]*workerState
}

func NewServer(logger *slog.Logger, metrics *metrics.Registry, store *storage.SQLiteStore, cat *catalog.Catalog) *Server {
	return &Server{logger: logger, metrics: metrics, store: store, cat: cat, workers: map[string]*workerState{}}
}

func (s *Server) RegisterWorker(args rpct.RegisterWorkerArgs, reply *rpct.RegisterWorkerReply) error {
	s.mu.Lock()
	s.workers[args.WorkerID] = &workerState{Address: args.Address, LastSeen: time.Now().UTC()}
	s.rebalanceMasterLocked(context.Background())
	s.mu.Unlock()
	reply.Accepted = true
	s.metrics.IncEvents()
	return nil
}

func (s *Server) Heartbeat(args rpct.HeartbeatArgs, reply *rpct.HeartbeatReply) error {
	s.mu.Lock()
	if st, ok := s.workers[args.WorkerID]; ok {
		st.LastSeen = time.Now().UTC()
		st.VirtualLoad = args.VirtualLoad
	}
	s.rebalanceMasterLocked(context.Background())
	s.mu.Unlock()
	reply.OK = true
	return nil
}

func (s *Server) PushScenario(args rpct.PushScenarioArgs, reply *rpct.PushScenarioReply) error {
	if _, err := scenario.ParseYAMLLike(bytes.NewBufferString(args.DSL)); err != nil {
		return fmt.Errorf("dsl parse failed: %w", err)
	}
	id := newID("scn")
	if err := s.store.SaveScenario(context.Background(), id, args.Name, args.DSL); err != nil {
		return err
	}
	reply.ScenarioID = id
	s.metrics.IncEvents()
	return nil
}

func (s *Server) RunExperiment(args rpct.RunExperimentArgs, reply *rpct.RunExperimentReply) error {
	dsl, err := s.store.GetScenario(context.Background(), args.ScenarioID)
	if err != nil {
		return err
	}
	scn, err := scenario.ParseYAMLLike(bytes.NewBufferString(dsl))
	if err != nil {
		return err
	}
	scn, err = s.cat.ResolveScenario(scn)
	if err != nil {
		return err
	}
	runID := newID("run")
	if err := s.dispatchToWorkers(context.Background(), runID, args.Seed, scn.Devices); err != nil {
		return err
	}
	reply.RunID = runID
	s.metrics.IncEvents()
	return nil
}

func (s *Server) dispatchToWorkers(ctx context.Context, runID string, seed int64, devices []scenario.Device) error {
	s.mu.Lock()
	s.rebalanceMasterLocked(ctx)
	active := s.activeWorkersLocked()
	s.mu.Unlock()
	if len(active) == 0 {
		return fmt.Errorf("no active workers registered")
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
		client, err := rpc.Dial("tcp", ws.Address)
		if err != nil {
			return fmt.Errorf("dial worker %s: %w", wid, err)
		}
		args := rpct.StartWorkloadArgs{RunID: runID, Seed: seed, Devices: buckets[wid], IsMaster: ws.IsMaster}
		var out rpct.StartWorkloadReply
		err = client.Call("WorkerService.StartWorkload", args, &out)
		_ = client.Close()
		if err != nil {
			return fmt.Errorf("start workload on %s: %w", wid, err)
		}
	}
	return nil
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
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		s.logger.Error("assign role dial failed", "worker", workerID, "err", err)
		return
	}
	defer client.Close()
	var out rpct.AssignRoleReply
	if err := client.Call("WorkerService.AssignRole", rpct.AssignRoleArgs{IsMaster: isMaster}, &out); err != nil {
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

func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
