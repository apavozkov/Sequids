package orchestratorpb

import (
	context "context"
	"errors"

	grpc "google.golang.org/grpc"
)

const (
	WorkerService_CreateSensor_FullMethodName             = "/orchestratorpb.WorkerService/CreateSensor"
	OrchestratorService_ReportSensorStatus_FullMethodName = "/orchestratorpb.OrchestratorService/ReportSensorStatus"
)

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

type WorkerServiceServer interface {
	CreateSensor(context.Context, *CreateSensorRequest) (*CreateSensorResponse, error)
	mustEmbedUnimplementedWorkerServiceServer()
}

type UnimplementedWorkerServiceServer struct{}

func (UnimplementedWorkerServiceServer) CreateSensor(context.Context, *CreateSensorRequest) (*CreateSensorResponse, error) {
	return nil, errors.New("method CreateSensor not implemented")
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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: WorkerService_CreateSensor_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).CreateSensor(ctx, req.(*CreateSensorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orchestratorpb.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods:     []grpc.MethodDesc{{MethodName: "CreateSensor", Handler: _WorkerService_CreateSensor_Handler}},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "orchestrator.proto",
}

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

type OrchestratorServiceServer interface {
	ReportSensorStatus(context.Context, *ReportSensorStatusRequest) (*ReportSensorStatusResponse, error)
	mustEmbedUnimplementedOrchestratorServiceServer()
}

type UnimplementedOrchestratorServiceServer struct{}

func (UnimplementedOrchestratorServiceServer) ReportSensorStatus(context.Context, *ReportSensorStatusRequest) (*ReportSensorStatusResponse, error) {
	return nil, errors.New("method ReportSensorStatus not implemented")
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
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: OrchestratorService_ReportSensorStatus_FullMethodName}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrchestratorServiceServer).ReportSensorStatus(ctx, req.(*ReportSensorStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var OrchestratorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "orchestratorpb.OrchestratorService",
	HandlerType: (*OrchestratorServiceServer)(nil),
	Methods:     []grpc.MethodDesc{{MethodName: "ReportSensorStatus", Handler: _OrchestratorService_ReportSensorStatus_Handler}},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "orchestrator.proto",
}
