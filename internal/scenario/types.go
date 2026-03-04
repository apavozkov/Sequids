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
}

type Anomaly struct {
	Kind        string
	Probability float64
	Amplitude   float64
	DriftPerSec float64
}
