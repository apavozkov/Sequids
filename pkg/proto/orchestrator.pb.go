package orchestratorpb

// NOTE: MVP uses JSON codec over gRPC for transport simplicity.
// Structs are kept compatible with proto field names for future protoc migration.

type CreateSensorRequest struct {
	SensorId   string `json:"sensor_id,omitempty"`
	IntervalMs int64  `json:"interval_ms,omitempty"`
	WorkerId   string `json:"worker_id,omitempty"`
}

func (x *CreateSensorRequest) GetSensorId() string {
	if x != nil {
		return x.SensorId
	}
	return ""
}

func (x *CreateSensorRequest) GetIntervalMs() int64 {
	if x != nil {
		return x.IntervalMs
	}
	return 0
}

func (x *CreateSensorRequest) GetWorkerId() string {
	if x != nil {
		return x.WorkerId
	}
	return ""
}

type CreateSensorResponse struct {
	Status string `json:"status,omitempty"`
}

func (x *CreateSensorResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type ReportSensorStatusRequest struct {
	SensorId string  `json:"sensor_id,omitempty"`
	Status   string  `json:"status,omitempty"`
	Value    float64 `json:"value,omitempty"`
}

func (x *ReportSensorStatusRequest) GetSensorId() string {
	if x != nil {
		return x.SensorId
	}
	return ""
}

func (x *ReportSensorStatusRequest) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *ReportSensorStatusRequest) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type ReportSensorStatusResponse struct {
	Ack string `json:"ack,omitempty"`
}

func (x *ReportSensorStatusResponse) GetAck() string {
	if x != nil {
		return x.Ack
	}
	return ""
}
