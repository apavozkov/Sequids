package scenario

type Scenario struct {
	Name    string
	Devices []Device
	Bridges []Bridge
	Flows   []Flow
}

type Device struct {
	ID          string
	Type        string
	Topic       string
	From        string
	FrequencyHz float64
	FormulaRef  string
	Formula     string
	AnomalyRefs []string
	Anomalies   []Anomaly

	Gain            float64
	Offset          float64
	ClampMin        *float64
	ClampMax        *float64
	StartupDelaySec float64
	JitterRatio     float64
}

type Anomaly struct {
	Kind        string
	Probability float64
	Amplitude   float64
	DriftPerSec float64
	DurationSec float64
	HoldSec     float64
}

type Bridge struct {
	ID           string
	Protocol     string
	Mode         string
	IngressTopic string
	EgressTopic  string
}

type Flow struct {
	ID         string
	Device     string
	Conditions []Condition
	Actions    []Action
}

type Condition struct {
	Metric     string
	Op         string
	Threshold  *float64
	Min        *float64
	Max        *float64
	SustainSec float64
}

type Action struct {
	Target       string
	Command      string
	PayloadField string
	CooldownSec  float64
}
