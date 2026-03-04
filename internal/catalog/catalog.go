package catalog

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sequids/internal/scenario"
)

type Formula struct {
	Name        string
	Description string
	Expression  string
}

type AnomalyProfile struct {
	Name        string
	Description string
	Anomaly     scenario.Anomaly
}

type Catalog struct {
	Formulas  map[string]Formula
	Anomalies map[string]AnomalyProfile
}

func Load(formulasPath, anomaliesPath string) (*Catalog, error) {
	f, err := loadFormulas(formulasPath)
	if err != nil {
		return nil, err
	}
	a, err := loadAnomalies(anomaliesPath)
	if err != nil {
		return nil, err
	}
	return &Catalog{Formulas: f, Anomalies: a}, nil
}

func loadFormulas(path string) (map[string]Formula, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]Formula{}
	var cur *Formula
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || line == "formulas:" {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			name := strings.TrimSuffix(line, ":")
			if cur != nil {
				out[cur.Name] = *cur
			}
			cur = &Formula{Name: name}
			continue
		}
		if cur == nil {
			continue
		}
		if strings.HasPrefix(line, "description:") {
			cur.Description = trim(strings.TrimPrefix(line, "description:"))
		} else if strings.HasPrefix(line, "expression:") {
			cur.Expression = trim(strings.TrimPrefix(line, "expression:"))
		}
	}
	if cur != nil {
		out[cur.Name] = *cur
	}
	return out, s.Err()
}

func loadAnomalies(path string) (map[string]AnomalyProfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]AnomalyProfile{}
	var cur *AnomalyProfile
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || line == "anomalies:" {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			name := strings.TrimSuffix(line, ":")
			if cur != nil {
				out[cur.Name] = *cur
			}
			cur = &AnomalyProfile{Name: name}
			continue
		}
		if cur == nil {
			continue
		}
		switch {
		case strings.HasPrefix(line, "description:"):
			cur.Description = trim(strings.TrimPrefix(line, "description:"))
		case strings.HasPrefix(line, "kind:"):
			cur.Anomaly.Kind = trim(strings.TrimPrefix(line, "kind:"))
		case strings.HasPrefix(line, "probability:"):
			v, _ := strconv.ParseFloat(trim(strings.TrimPrefix(line, "probability:")), 64)
			cur.Anomaly.Probability = v
		case strings.HasPrefix(line, "amplitude:"):
			v, _ := strconv.ParseFloat(trim(strings.TrimPrefix(line, "amplitude:")), 64)
			cur.Anomaly.Amplitude = v
		case strings.HasPrefix(line, "drift_per_sec:"):
			v, _ := strconv.ParseFloat(trim(strings.TrimPrefix(line, "drift_per_sec:")), 64)
			cur.Anomaly.DriftPerSec = v
		}
	}
	if cur != nil {
		out[cur.Name] = *cur
	}
	return out, s.Err()
}

func trim(s string) string {
	return strings.Trim(strings.TrimSpace(s), "\"")
}

func (c *Catalog) ResolveScenario(in scenario.Scenario) (scenario.Scenario, error) {
	out := in
	for di := range out.Devices {
		d := &out.Devices[di]
		if d.FormulaRef != "" {
			f, ok := c.Formulas[d.FormulaRef]
			if !ok {
				return scenario.Scenario{}, fmt.Errorf("unknown formula_ref: %s", d.FormulaRef)
			}
			d.Formula = f.Expression
		}
		if len(d.AnomalyRefs) > 0 {
			d.Anomalies = nil
			for _, ref := range d.AnomalyRefs {
				p, ok := c.Anomalies[ref]
				if !ok {
					return scenario.Scenario{}, fmt.Errorf("unknown anomaly_ref: %s", ref)
				}
				d.Anomalies = append(d.Anomalies, p.Anomaly)
			}
		}
	}
	return out, nil
}
