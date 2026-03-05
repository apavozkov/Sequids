package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"sequids/internal/scenario"
)

func TestResolveScenario(t *testing.T) {
	d := t.TempDir()
	fp := filepath.Join(d, "formulas.yaml")
	ap := filepath.Join(d, "anomalies.yaml")
	_ = os.WriteFile(fp, []byte("formulas:\n  f1:\n    description: \"x\"\n    expression: \"10+t\"\n"), 0o644)
	_ = os.WriteFile(ap, []byte("anomalies:\n  a1:\n    description: \"n\"\n    kind: noise\n    probability: 0.1\n    amplitude: 2\n"), 0o644)
	cat, err := Load(fp, ap)
	if err != nil {
		t.Fatal(err)
	}
	out, err := cat.ResolveScenario(scenario.Scenario{Name: "x", Devices: []scenario.Device{{ID: "d1", FormulaRef: "f1", AnomalyRefs: []string{"a1"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Devices[0].Formula != "10+t" || len(out.Devices[0].Anomalies) != 1 {
		t.Fatalf("bad resolution: %+v", out)
	}
}
