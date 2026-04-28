package scenario

import (
	"strings"
	"testing"
)

func TestParseYAMLLike(t *testing.T) {
	dsl := `name: demo
devices:
  - id: d1
    type: temp
    topic: a/b
    frequency_hz: 1
    formula_ref: temp_daily_sine
    anomalies:
      - anomaly_ref: mild_sensor_noise`
	s, err := ParseYAMLLike(strings.NewReader(dsl))
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "demo" || len(s.Devices) != 1 || s.Devices[0].FormulaRef != "temp_daily_sine" || len(s.Devices[0].AnomalyRefs) != 1 {
		t.Fatalf("unexpected parse result: %+v", s)
	}
}

func TestParseYAMLLike_FlowsAndBridges(t *testing.T) {
	dsl := `name: bus-demo
devices:
  - id: temp-1
    type: temperature
    topic: iot/temp
bridges:
  - id: b1
    protocol: mqtt
    mode: bus_gateway
    ingress_topic: iot/in
    egress_topic: iot/out
flows:
  - id: f1
    from: temp-1
    to: ac-1
    via: adapter.bus
    bridge_ref: b1
    conditions:
      - metric: value
        op: gt
        threshold: 26
    actions:
      - command: power_on
        target: ac-1`
	s, err := ParseYAMLLike(strings.NewReader(dsl))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Bridges) != 1 || s.Bridges[0].ID != "b1" {
		t.Fatalf("expected 1 bridge b1, got %+v", s.Bridges)
	}
	if len(s.Flows) != 1 || s.Flows[0].ID != "f1" || len(s.Flows[0].Conditions) != 1 || len(s.Flows[0].Actions) != 1 {
		t.Fatalf("unexpected flow parse result: %+v", s.Flows)
	}
}
