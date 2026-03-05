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
