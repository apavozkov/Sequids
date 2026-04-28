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
	var currentBridge *Bridge
	var currentFlow *Flow
	var currentCondition *Condition
	var currentAction *Action
	section := ""

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
	flushCondition := func() {
		if currentFlow != nil && currentCondition != nil {
			currentFlow.Conditions = append(currentFlow.Conditions, *currentCondition)
			currentCondition = nil
		}
	}
	flushAction := func() {
		if currentFlow != nil && currentAction != nil {
			currentFlow.Actions = append(currentFlow.Actions, *currentAction)
			currentAction = nil
		}
	}
	flushFlow := func() {
		flushCondition()
		flushAction()
		if currentFlow != nil {
			s.Flows = append(s.Flows, *currentFlow)
			currentFlow = nil
		}
	}
	flushBridge := func() {
		if currentBridge != nil {
			s.Bridges = append(s.Bridges, *currentBridge)
			currentBridge = nil
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
			section = line
			continue
		case line == "bridges:":
			flushDevice()
			flushFlow()
			section = "bridges:"
			continue
		case line == "flows:":
			flushDevice()
			flushBridge()
			section = "flows:"
			continue
		case line == "conditions:":
			section = "conditions:"
			continue
		case line == "actions:":
			section = "actions:"
			continue
		case strings.HasPrefix(line, "- id:"):
			switch section {
			case "devices:", "anomalies:":
				flushDevice()
				currentDevice = &Device{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
			case "bridges:":
				flushBridge()
				currentBridge = &Bridge{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
			case "flows:":
				flushFlow()
				currentFlow = &Flow{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
			}
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
		case strings.HasPrefix(line, "protocol:") && currentBridge != nil:
			currentBridge.Protocol = strings.TrimSpace(strings.TrimPrefix(line, "protocol:"))
		case strings.HasPrefix(line, "mode:") && currentBridge != nil:
			currentBridge.Mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
		case strings.HasPrefix(line, "ingress_topic:") && currentBridge != nil:
			currentBridge.IngressTopic = strings.TrimSpace(strings.TrimPrefix(line, "ingress_topic:"))
		case strings.HasPrefix(line, "egress_topic:") && currentBridge != nil:
			currentBridge.EgressTopic = strings.TrimSpace(strings.TrimPrefix(line, "egress_topic:"))
		case strings.HasPrefix(line, "from:") && currentFlow != nil:
			currentFlow.From = strings.TrimSpace(strings.TrimPrefix(line, "from:"))
		case strings.HasPrefix(line, "to:") && currentFlow != nil:
			currentFlow.To = strings.TrimSpace(strings.TrimPrefix(line, "to:"))
		case strings.HasPrefix(line, "via:") && currentFlow != nil:
			currentFlow.Via = strings.TrimSpace(strings.TrimPrefix(line, "via:"))
		case strings.HasPrefix(line, "bridge_ref:") && currentFlow != nil:
			currentFlow.BridgeRef = strings.TrimSpace(strings.TrimPrefix(line, "bridge_ref:"))
		case strings.HasPrefix(line, "- metric:") && currentFlow != nil:
			flushCondition()
			currentCondition = &Condition{Metric: strings.TrimSpace(strings.TrimPrefix(line, "- metric:"))}
		case strings.HasPrefix(line, "metric:") && currentCondition != nil:
			currentCondition.Metric = strings.TrimSpace(strings.TrimPrefix(line, "metric:"))
		case strings.HasPrefix(line, "op:") && currentCondition != nil:
			currentCondition.Op = strings.TrimSpace(strings.TrimPrefix(line, "op:"))
		case strings.HasPrefix(line, "threshold:") && currentCondition != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "threshold:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad threshold: %w", lineNo, err)
			}
			currentCondition.Threshold = &v
		case strings.HasPrefix(line, "min:") && currentCondition != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "min:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad min: %w", lineNo, err)
			}
			currentCondition.Min = &v
		case strings.HasPrefix(line, "max:") && currentCondition != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "max:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad max: %w", lineNo, err)
			}
			currentCondition.Max = &v
		case strings.HasPrefix(line, "sustain_sec:") && currentCondition != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "sustain_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad sustain_sec: %w", lineNo, err)
			}
			currentCondition.SustainSec = v
		case strings.HasPrefix(line, "- command:") && currentFlow != nil:
			flushAction()
			currentAction = &Action{Command: strings.TrimSpace(strings.TrimPrefix(line, "- command:"))}
		case strings.HasPrefix(line, "command:") && currentAction != nil:
			currentAction.Command = strings.TrimSpace(strings.TrimPrefix(line, "command:"))
		case strings.HasPrefix(line, "target:") && currentAction != nil:
			currentAction.Target = strings.TrimSpace(strings.TrimPrefix(line, "target:"))
		case strings.HasPrefix(line, "payload_field:") && currentAction != nil:
			currentAction.PayloadField = strings.TrimSpace(strings.TrimPrefix(line, "payload_field:"))
		case strings.HasPrefix(line, "cooldown_sec:") && currentAction != nil:
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "cooldown_sec:")), 64)
			if err != nil {
				return Scenario{}, fmt.Errorf("line %d: bad cooldown_sec: %w", lineNo, err)
			}
			currentAction.CooldownSec = v
		}
	}
	if err := scanner.Err(); err != nil {
		return Scenario{}, err
	}
	flushDevice()
	flushBridge()
	flushFlow()
	if s.Name == "" {
		return Scenario{}, fmt.Errorf("scenario name is required")
	}
	return s, nil
}
