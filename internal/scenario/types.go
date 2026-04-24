package scenario

type Scenario struct {
	Name    string
	Devices []Device
}

type Device struct {
	ID          string
	Type        string
	Topic       string
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
