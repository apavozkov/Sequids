package orchestrator

import (
	"sync"
	"time"
)

type Scenario struct {
	SensorID   string
	WorkerID   string
	Interval   time.Duration
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LastStatus string
	LastValue  float64
}

type Store struct {
	mu        sync.RWMutex
	scenarios map[string]*Scenario
}

func NewStore() *Store {
	return &Store{scenarios: make(map[string]*Scenario)}
}

func (s *Store) UpsertScenario(scenario *Scenario) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scenarios[scenario.SensorID] = scenario
}

func (s *Store) UpdateStatus(sensorID, status string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	scenario, ok := s.scenarios[sensorID]
	if !ok {
		return
	}
	scenario.LastStatus = status
	scenario.LastValue = value
	scenario.UpdatedAt = time.Now()
}

func (s *Store) GetScenario(sensorID string) (*Scenario, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scenario, ok := s.scenarios[sensorID]
	return scenario, ok
}
