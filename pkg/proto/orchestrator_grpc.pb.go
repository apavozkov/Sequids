// Code generated manually for the MVP; mimics protoc-gen-go-grpc output.
// source: orchestrator.proto

package orchestratorpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	WorkerService_CreateSensor_FullMethodName = "/orchestratorpb.WorkerService/CreateSensor"
	OrchestratorService_ReportSensorStatus_FullMethodName = "/orchestratorpb.OrchestratorService/ReportSensorStatus"
)

// WorkerServiceClient is the client API for WorkerService service.
type WorkerServiceClient interface {
	CreateSensor(ctx context.Context, in *CreateSensorRequest, opts ...grpc.CallOption) (*CreateSensorResponse, error)
}

type workerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerServiceClient(cc grpc.ClientConnInterface) WorkerServiceClient {
	return &workerServiceClient{cc}
}

func (c *workerServiceClient) CreateSensor(ctx context.Context, in *CreateSensorRequest, opts ...grpc.CallOption) (*CreateSensorResponse, error) {
	out := new(CreateSensorResponse)
	err := c.cc.Invoke(ctx, WorkerService_CreateSensor_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerServiceServer is the server API for WorkerService service.
type WorkerServiceServer interface {
	CreateSensor(context.Context, *CreateSensorRequest) (*CreateSensorResponse, error)
	mustEmbedUnimplementedWorkerServiceServer()
}

// UnimplementedWorkerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWorkerServiceServer struct{}

func (UnimplementedWorkerServiceServer) CreateSensor(context.Context, *CreateSensorRequest) (*CreateSensorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSensor not implemented")
}
func (UnimplementedWorkerServiceServer) mustEmbedUnimplementedWorkerServiceServer() {}

func RegisterWorkerServiceServer(s grpc.ServiceRegistrar, srv WorkerServiceServer) {
	s.RegisterService(&WorkerService_ServiceDesc, srv)
}

func _WorkerService_CreateSensor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSensorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).CreateSensor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_CreateSensor_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).CreateSensor(ctx, req.(*CreateSensorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorkerService_ServiceDesc is the grpc.ServiceDesc for WorkerService service.
var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orchestratorpb.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateSensor",
			Handler:    _WorkerService_CreateSensor_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orchestrator.proto",
}

// OrchestratorServiceClient is the client API for OrchestratorService service.
type OrchestratorServiceClient interface {
	ReportSensorStatus(ctx context.Context, in *ReportSensorStatusRequest, opts ...grpc.CallOption) (*ReportSensorStatusResponse, error)
}

type orchestratorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOrchestratorServiceClient(cc grpc.ClientConnInterface) OrchestratorServiceClient {
	return &orchestratorServiceClient{cc}
}

func (c *orchestratorServiceClient) ReportSensorStatus(ctx context.Context, in *ReportSensorStatusRequest, opts ...grpc.CallOption) (*ReportSensorStatusResponse, error) {
	out := new(ReportSensorStatusResponse)
	err := c.cc.Invoke(ctx, OrchestratorService_ReportSensorStatus_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrchestratorServiceServer is the server API for OrchestratorService service.
type OrchestratorServiceServer interface {
	ReportSensorStatus(context.Context, *ReportSensorStatusRequest) (*ReportSensorStatusResponse, error)
	mustEmbedUnimplementedOrchestratorServiceServer()
}

// UnimplementedOrchestratorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedOrchestratorServiceServer struct{}

func (UnimplementedOrchestratorServiceServer) ReportSensorStatus(context.Context, *ReportSensorStatusRequest) (*ReportSensorStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportSensorStatus not implemented")
}
func (UnimplementedOrchestratorServiceServer) mustEmbedUnimplementedOrchestratorServiceServer() {}

func RegisterOrchestratorServiceServer(s grpc.ServiceRegistrar, srv OrchestratorServiceServer) {
	s.RegisterService(&OrchestratorService_ServiceDesc, srv)
}

func _OrchestratorService_ReportSensorStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportSensorStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrchestratorServiceServer).ReportSensorStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OrchestratorService_ReportSensorStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrchestratorServiceServer).ReportSensorStatus(ctx, req.(*ReportSensorStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OrchestratorService_ServiceDesc is the grpc.ServiceDesc for OrchestratorService service.
var OrchestratorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orchestratorpb.OrchestratorService",
	HandlerType: (*OrchestratorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReportSensorStatus",
			Handler:    _OrchestratorService_ReportSensorStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orchestrator.proto",
}

