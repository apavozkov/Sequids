package scenario

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseYAMLLike parses a minimal YAML-like DSL tailored for IoT scenarios.
func ParseYAMLLike(r io.Reader) (Scenario, error) {
	s := Scenario{}
	var currentDevice *Device
	var currentInline *Anomaly

	flushInline := func() {
		if currentDevice != nil && currentInline != nil {
			currentDevice.Anomalies = append(currentDevice.Anomalies, *currentInline)
			currentInline = nil
		}
	}
	flushDevice := func() {
		flushInline()
		if currentDevice != nil {
			s.Devices = append(s.Devices, *currentDevice)
			currentDevice = nil
		}
	}

	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "name:"):
			s.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		case line == "devices:" || line == "anomalies:":
			continue
		case strings.HasPrefix(line, "- id:"):
			flushDevice()
			currentDevice = &Device{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
		case strings.HasPrefix(line, "type:") && currentDevice != nil:
			currentDevice.Type = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
		case strings.HasPrefix(line, "topic:") && currentDevice != nil:
			currentDevice.Topic = strings.TrimSpace(strings.TrimPrefix(line, "topic:"))
		case strings.HasPrefix(line, "frequency_hz:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "frequency_hz:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad frequency_hz: %w", lineNo, err)
			}
			currentDevice.FrequencyHz = v
		case strings.HasPrefix(line, "gain:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "gain:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad gain: %w", lineNo, err)
			}
			currentDevice.Gain = v
		case strings.HasPrefix(line, "offset:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "offset:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad offset: %w", lineNo, err)
			}
			currentDevice.Offset = v
		case strings.HasPrefix(line, "clamp_min:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "clamp_min:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad clamp_min: %w", lineNo, err)
			}
			currentDevice.ClampMin = &v
		case strings.HasPrefix(line, "clamp_max:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "clamp_max:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad clamp_max: %w", lineNo, err)
			}
			currentDevice.ClampMax = &v
		case strings.HasPrefix(line, "startup_delay_sec:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "startup_delay_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad startup_delay_sec: %w", lineNo, err)
			}
			currentDevice.StartupDelaySec = v
		case strings.HasPrefix(line, "jitter_ratio:") && currentDevice != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "jitter_ratio:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad jitter_ratio: %w", lineNo, err)
			}
			currentDevice.JitterRatio = v
		case strings.HasPrefix(line, "formula_ref:") && currentDevice != nil:
			currentDevice.FormulaRef = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "formula_ref:")), "\"")
		case strings.HasPrefix(line, "- anomaly_ref:") && currentDevice != nil:
			flushInline()
			currentDevice.AnomalyRefs = append(currentDevice.AnomalyRefs, strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "- anomaly_ref:")), "\""))
		case strings.HasPrefix(line, "- kind:") && currentDevice != nil:
			flushInline()
			currentInline = &Anomaly{Kind: strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "- kind:")), "\"")}
		case strings.HasPrefix(line, "kind:") && currentInline != nil:
			currentInline.Kind = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "kind:")), "\"")
		case strings.HasPrefix(line, "probability:") && currentInline != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "probability:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad probability: %w", lineNo, err)
			}
			currentInline.Probability = v
		case strings.HasPrefix(line, "amplitude:") && currentInline != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "amplitude:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad amplitude: %w", lineNo, err)
			}
			currentInline.Amplitude = v
		case strings.HasPrefix(line, "drift_per_sec:") && currentInline != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "drift_per_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad drift_per_sec: %w", lineNo, err)
			}
			currentInline.DriftPerSec = v
		case strings.HasPrefix(line, "duration_sec:") && currentInline != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "duration_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad duration_sec: %w", lineNo, err)
			}
			currentInline.DurationSec = v
		case strings.HasPrefix(line, "hold_sec:") && currentInline != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "hold_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad hold_sec: %w", lineNo, err)
			}
			currentInline.HoldSec = v
		}
	}
	if err := scanner.Err(); err != nil {
		return Scenario{}, err
	}
	flushDevice()
	if s.Name == "" {
		return Scenario{}, fmt.Errorf("scenario name is required")
	}
	return s, nil
}
