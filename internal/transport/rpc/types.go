package rpc

import "sequids/internal/scenario"

type RegisterWorkerArgs struct {
	WorkerID string
	Address  string
}

type RegisterWorkerReply struct{ Accepted bool }

type HeartbeatArgs struct {
	WorkerID    string
	VirtualLoad int
}

type HeartbeatReply struct{ OK bool }

type PushScenarioArgs struct {
	Name string
	DSL  string
}

type PushScenarioReply struct{ ScenarioID string }

type RunExperimentArgs struct {
	ScenarioID string
	Seed       int64
}

type RunExperimentReply struct{ RunID string }

type AssignRoleArgs struct{ IsMaster bool }
type AssignRoleReply struct{ OK bool }

type StartWorkloadArgs struct {
	RunID    string
	Seed     int64
	Devices  []scenario.Device
	IsMaster bool
}

type StartWorkloadReply struct{ Started bool }
