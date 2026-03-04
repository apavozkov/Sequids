package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
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
	RunID    string
	DeviceID string
	Topic    string
	Value    float64
	Source   string // virtual|real
	TS       time.Time
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
}

func NewRuntime(logger *slog.Logger, m *metrics.Registry, pub mqtt.Publisher, influxURL, influxToken, org, bucket string) *Runtime {
	r := &Runtime{logger: logger, metrics: m, pub: pub, influxURL: influxURL, influxToken: influxToken, org: org, bucket: bucket, bus: make(chan DataPoint, 4096)}
	go r.listenerLoop(context.Background())
	return r
}

func (r *Runtime) SetMaster(v bool) { r.isMaster.Store(v) }
func (r *Runtime) VirtualLoad() int { return int(r.load.Load()) }

func (r *Runtime) Start(ctx context.Context, runID string, seed int64, devices []scenario.Device) {
	var wg sync.WaitGroup
	for idx, device := range devices {
		d := device
		wg.Add(1)
		go func(offset int64) {
			defer wg.Done()
			r.runDevice(ctx, rand.New(rand.NewSource(seed+offset+1)), runID, d)
		}(int64(idx))
	}
	go func() { wg.Wait() }()
}

func (r *Runtime) runDevice(ctx context.Context, rng *rand.Rand, runID string, d scenario.Device) {
	r.load.Add(1)
	defer r.load.Add(-1)
	if d.FrequencyHz <= 0 {
		d.FrequencyHz = 1
	}
	ticker := time.NewTicker(time.Duration(float64(time.Second) / d.FrequencyHz))
	defer ticker.Stop()
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t := time.Since(start).Seconds()
			value := evalFormula(d.Formula, t)
			for _, a := range d.Anomalies {
				if rng.Float64() <= a.Probability {
					switch a.Kind {
					case "noise", "false_data":
						value += (rng.Float64()*2 - 1) * a.Amplitude
					case "drift":
						value += t * a.DriftPerSec
					}
				}
			}
			msg := fmt.Sprintf(`{"run_id":"%s","device_id":"%s","value":%f,"ts":"%s"}`,
				runID, d.ID, value, time.Now().UTC().Format(time.RFC3339Nano))
			if err := r.pub.Publish(ctx, d.Topic, []byte(msg)); err != nil {
				r.metrics.IncErrors()
				r.logger.Error("publish failed", "device", d.ID, "err", err)
				continue
			}
			r.metrics.IncEvents()
			r.enqueue(DataPoint{RunID: runID, DeviceID: d.ID, Topic: d.Topic, Value: value, Source: "virtual", TS: time.Now().UTC()})
		}
	}
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
			if p.Source == "real" && !r.isMaster.Load() {
				continue
			}
			if err := r.writeInflux(p); err != nil {
				r.metrics.IncErrors()
				r.logger.Error("influx write failed", "err", err)
			}
		}
	}
}

func (r *Runtime) writeInflux(p DataPoint) error {
	if r.influxURL == "" || r.influxToken == "" {
		return nil
	}
	line := fmt.Sprintf("device_metrics,device_id=%s,topic=%s,source=%s value=%f %d",
		escapeTag(p.DeviceID), escapeTag(p.Topic), escapeTag(p.Source), p.Value, p.TS.UnixNano())
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
