package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sequids/internal/adapters/mqtt"
	"sequids/internal/metrics"
	"sequids/internal/scenario"
)

type DataPoint struct {
	RunID      string
	DeviceID   string
	DeviceType string
	Topic      string
	Value      float64
	Source     string // virtual|real
	TS         time.Time
	Payload    []byte
}

type runState struct {
	cancel  func()
	devices int
}

type Runtime struct {
	logger      *slog.Logger
	metrics     *metrics.Registry
	pub         mqtt.Publisher
	influxURL   string
	influxToken string
	org         string
	bucket      string

	bus      chan DataPoint
	isMaster atomic.Bool
	load     atomic.Int64

	mu   sync.RWMutex
	runs map[string]runState

	topicMu     sync.RWMutex
	topicValues map[string]float64
}

func NewRuntime(logger *slog.Logger, m *metrics.Registry, pub mqtt.Publisher, influxURL, influxToken, org, bucket string) *Runtime {
	r := &Runtime{logger: logger, metrics: m, pub: pub, influxURL: influxURL, influxToken: influxToken, org: org, bucket: bucket, bus: make(chan DataPoint, 4096), runs: map[string]runState{}, topicValues: map[string]float64{}}
	go r.listenerLoop(context.Background())
	return r
}

func (r *Runtime) SetMaster(v bool) { r.isMaster.Store(v) }
func (r *Runtime) VirtualLoad() int { return int(r.load.Load()) }

func (r *Runtime) Status() (activeRuns, activeDevices int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, st := range r.runs {
		activeDevices += st.devices
	}
	return len(r.runs), activeDevices
}

func (r *Runtime) Start(ctx context.Context, runID string, seed int64, devices []scenario.Device) {
	r.Stop(runID)
	runCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.runs[runID] = runState{cancel: cancel, devices: len(devices)}
	r.mu.Unlock()

	for idx, device := range devices {
		d := device
		go func(offset int64) {
			r.runDevice(runCtx, rand.New(rand.NewSource(seed+offset+1)), runID, d)
		}(int64(idx))
	}
}

func (r *Runtime) Stop(runID string) bool {
	r.mu.Lock()
	st, ok := r.runs[runID]
	if ok {
		delete(r.runs, runID)
	}
	r.mu.Unlock()
	if ok {
		st.cancel()
	}
	return ok
}

func (r *Runtime) runDevice(ctx context.Context, rng *rand.Rand, runID string, d scenario.Device) {
	if strings.HasPrefix(d.Formula, "control:") {
		r.runControlledDevice(ctx, runID, d)
		return
	}
	r.load.Add(1)
	defer r.load.Add(-1)
	defer r.cleanupRun(runID)

	if d.FrequencyHz <= 0 {
		d.FrequencyHz = 1
	}
	if d.Gain == 0 {
		d.Gain = 1
	}
	if d.StartupDelaySec > 0 {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(d.StartupDelaySec * float64(time.Second))):
		}
	}

	start := time.Now()
	var stuckUntil time.Time
	var stuckValue float64

	for {
		wait := time.Duration(float64(time.Second) / d.FrequencyHz)
		if d.JitterRatio > 0 {
			jitter := 1 + (rng.Float64()*2-1)*d.JitterRatio
			if jitter < 0.1 {
				jitter = 0.1
			}
			wait = time.Duration(float64(wait) * jitter)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		t := time.Since(start).Seconds()
		value := evalFormula(d.Formula, t)*d.Gain + d.Offset
		if d.ClampMin != nil {
			value = math.Max(value, *d.ClampMin)
		}
		if d.ClampMax != nil {
			value = math.Min(value, *d.ClampMax)
		}

		now := time.Now()
		if now.Before(stuckUntil) {
			value = stuckValue
		}
		skipPublish := false
		delaySec := 0.0
		for _, a := range d.Anomalies {
			if rng.Float64() > a.Probability {
				continue
			}
			switch a.Kind {
			case "noise", "false_data", "spike":
				value += (rng.Float64()*2 - 1) * a.Amplitude
			case "drift":
				value += t * a.DriftPerSec
			case "dropout":
				skipPublish = true
			case "stuck":
				if a.HoldSec > 0 {
					stuckUntil = now.Add(time.Duration(a.HoldSec * float64(time.Second)))
					stuckValue = value
				}
			case "delay":
				if a.DurationSec > delaySec {
					delaySec = a.DurationSec
				}
			}
		}
		if skipPublish {
			continue
		}
		if delaySec > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(delaySec * float64(time.Second))):
			}
		}

		now = time.Now().UTC()
		msg := fmt.Sprintf(`{"run_id":"%s","device_id":"%s","device_type":"%s","value":%f,"state_code":%d,"state_text":"%s","ts":"%s"}`,
			runID, d.ID, d.Type, value, stateCode(d.Type, value), stateText(d.Type, value), now.Format(time.RFC3339Nano))
		r.enqueue(DataPoint{RunID: runID, DeviceID: d.ID, DeviceType: d.Type, Topic: d.Topic, Value: value, Source: "virtual", TS: now, Payload: []byte(msg)})
	}
}

func (r *Runtime) runControlledDevice(ctx context.Context, runID string, d scenario.Device) {
	r.load.Add(1)
	defer r.load.Add(-1)
	defer r.cleanupRun(runID)
	src, onGT, offLT, ok := parseControlFormula(d.Formula)
	if !ok {
		return
	}
	if d.FrequencyHz <= 0 {
		d.FrequencyHz = 1
	}
	isOn := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(float64(time.Second) / d.FrequencyHz)):
		}
		r.topicMu.RLock()
		v, exists := r.topicValues[src]
		r.topicMu.RUnlock()
		if !exists {
			continue
		}
		if v > onGT {
			isOn = true
		}
		if v < offLT {
			isOn = false
		}
		value := 0.0
		if isOn {
			value = 1
		}
		now := time.Now().UTC()
		msg := fmt.Sprintf(`{"run_id":"%s","device_id":"%s","device_type":"%s","value":%f,"state_code":%d,"state_text":"%s","source_value":%f,"ts":"%s"}`,
			runID, d.ID, d.Type, value, stateCode(d.Type, value), stateText(d.Type, value), v, now.Format(time.RFC3339Nano))
		r.enqueue(DataPoint{RunID: runID, DeviceID: d.ID, DeviceType: d.Type, Topic: d.Topic, Value: value, Source: "virtual", TS: now, Payload: []byte(msg)})
	}
}

func parseControlFormula(f string) (src string, onGT float64, offLT float64, ok bool) {
	if !strings.HasPrefix(f, "control:") {
		return "", 0, 0, false
	}
	rest := strings.TrimPrefix(f, "control:")
	parts := strings.Split(rest, ";")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "from":
			src = kv[1]
		case "on_gt":
			fmt.Sscanf(kv[1], "%f", &onGT)
		case "off_lt":
			fmt.Sscanf(kv[1], "%f", &offLT)
		}
	}
	return src, onGT, offLT, src != ""
}

func (r *Runtime) cleanupRun(runID string) {
	r.mu.Lock()
	st, ok := r.runs[runID]
	if ok {
		st.devices--
		if st.devices <= 0 {
			delete(r.runs, runID)
		} else {
			r.runs[runID] = st
		}
	}
	r.mu.Unlock()
}

func (r *Runtime) enqueue(p DataPoint) {
	select {
	case r.bus <- p:
	default:
		r.metrics.IncErrors()
	}
}

func (r *Runtime) listenerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-r.bus:
			r.topicMu.Lock()
			r.topicValues[p.Topic] = p.Value
			r.topicMu.Unlock()
			if p.Source == "real" && !r.isMaster.Load() {
				continue
			}
			if p.Source == "virtual" {
				payload := p.Payload
				if len(payload) == 0 {
					payload = []byte(fmt.Sprintf(`{"run_id":"%s","device_id":"%s","device_type":"%s","value":%f,"state_code":%d,"state_text":"%s","ts":"%s"}`, p.RunID, p.DeviceID, p.DeviceType, p.Value, stateCode(p.DeviceType, p.Value), stateText(p.DeviceType, p.Value), p.TS.Format(time.RFC3339Nano)))
				}
				if err := r.pub.Publish(ctx, p.Topic, payload); err != nil {
					r.metrics.IncErrors()
					r.logger.Error("publish failed", "device", p.DeviceID, "err", err)
					continue
				}
			}
			r.metrics.IncEvents()
			if err := r.writeInflux(p); err != nil {
				r.metrics.IncErrors()
				r.logger.Error("influx write failed", "err", err)
			}
		}
	}
}

func stateCode(deviceType string, value float64) int {
	switch deviceType {
	case "hvac", "ac", "air_conditioner":
		if value > 0 {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func stateText(deviceType string, value float64) string {
	if stateCode(deviceType, value) == 1 {
		return "on"
	}
	return "off"
}
func (r *Runtime) writeInflux(p DataPoint) error {
	if r.influxURL == "" || r.influxToken == "" {
		return nil
	}
	line := fmt.Sprintf("device_metrics,run_id=%s,device_id=%s,device_type=%s,topic=%s,source=%s value=%f %d",
		escapeTag(p.RunID), escapeTag(p.DeviceID), escapeTag(p.DeviceType), escapeTag(p.Topic), escapeTag(p.Source), p.Value, p.TS.UnixNano())
	if p.DeviceType == "hvac" || p.DeviceType == "ac" || p.DeviceType == "air_conditioner" {
		line += "\n" + fmt.Sprintf("device_metrics,run_id=%s,device_id=%s,device_type=%s,topic=%s,source=%s state_code=%di %d",
			escapeTag(p.RunID), escapeTag(p.DeviceID), escapeTag(p.DeviceType), escapeTag(p.Topic), escapeTag(p.Source), stateCode(p.DeviceType, p.Value), p.TS.UnixNano())
	}
	endpoint := fmt.Sprintf("%s/api/v2/write?org=%s&bucket=%s&precision=ns", strings.TrimRight(r.influxURL, "/"), r.org, r.bucket)
	req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(line))
	req.Header.Set("Authorization", "Token "+r.influxToken)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b))
	}
	return nil
}

func escapeTag(v string) string {
	return strings.NewReplacer(",", "\\,", " ", "\\ ", "=", "\\=").Replace(v)
}
