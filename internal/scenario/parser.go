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

	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "name:"):
			s.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		case line == "devices:":
			continue
		case strings.HasPrefix(line, "- id:"):
			if currentDevice != nil {
				s.Devices = append(s.Devices, *currentDevice)
			}
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
		case strings.HasPrefix(line, "formula_ref:") && currentDevice != nil:
			currentDevice.FormulaRef = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "formula_ref:")), "\"")
		case strings.HasPrefix(line, "- anomaly_ref:") && currentDevice != nil:
			currentDevice.AnomalyRefs = append(currentDevice.AnomalyRefs, strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "- anomaly_ref:")), "\""))
		}
	}
	if err := scanner.Err(); err != nil {
		return Scenario{}, err
	}
	if currentDevice != nil {
		s.Devices = append(s.Devices, *currentDevice)
	}
	if s.Name == "" {
		return Scenario{}, fmt.Errorf("scenario name is required")
	}
	return s, nil
}
