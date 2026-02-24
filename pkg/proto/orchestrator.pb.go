// Code generated manually for the MVP; mimics protoc-gen-go output.
// source: orchestrator.proto

package orchestratorpb

import (
	reflect "reflect"
	sync "sync"

	proto "google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Deprecated: Use CreateSensorRequest.ProtoReflect.Descriptor instead.
type CreateSensorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SensorId   string `protobuf:"bytes,1,opt,name=sensor_id,json=sensorId,proto3" json:"sensor_id,omitempty"`
	IntervalMs int64  `protobuf:"varint,2,opt,name=interval_ms,json=intervalMs,proto3" json:"interval_ms,omitempty"`
	WorkerId   string `protobuf:"bytes,3,opt,name=worker_id,json=workerId,proto3" json:"worker_id,omitempty"`
}

func (x *CreateSensorRequest) Reset() {
	*x = CreateSensorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_orchestrator_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateSensorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSensorRequest) ProtoMessage() {}

func (x *CreateSensorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_orchestrator_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSensorRequest.ProtoReflect.Descriptor instead.
func (*CreateSensorRequest) Descriptor() ([]byte, []int) {
	return file_orchestrator_proto_rawDescGZIP(), []int{0}
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

// Deprecated: Use CreateSensorResponse.ProtoReflect.Descriptor instead.
type CreateSensorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *CreateSensorResponse) Reset() {
	*x = CreateSensorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_orchestrator_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateSensorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSensorResponse) ProtoMessage() {}

func (x *CreateSensorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_orchestrator_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSensorResponse.ProtoReflect.Descriptor instead.
func (*CreateSensorResponse) Descriptor() ([]byte, []int) {
	return file_orchestrator_proto_rawDescGZIP(), []int{1}
}

func (x *CreateSensorResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

// Deprecated: Use ReportSensorStatusRequest.ProtoReflect.Descriptor instead.
type ReportSensorStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SensorId string  `protobuf:"bytes,1,opt,name=sensor_id,json=sensorId,proto3" json:"sensor_id,omitempty"`
	Status   string  `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Value    float64 `protobuf:"fixed64,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ReportSensorStatusRequest) Reset() {
	*x = ReportSensorStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_orchestrator_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReportSensorStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportSensorStatusRequest) ProtoMessage() {}

func (x *ReportSensorStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_orchestrator_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportSensorStatusRequest.ProtoReflect.Descriptor instead.
func (*ReportSensorStatusRequest) Descriptor() ([]byte, []int) {
	return file_orchestrator_proto_rawDescGZIP(), []int{2}
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

// Deprecated: Use ReportSensorStatusResponse.ProtoReflect.Descriptor instead.
type ReportSensorStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ack string `protobuf:"bytes,1,opt,name=ack,proto3" json:"ack,omitempty"`
}

func (x *ReportSensorStatusResponse) Reset() {
	*x = ReportSensorStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_orchestrator_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReportSensorStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReportSensorStatusResponse) ProtoMessage() {}

func (x *ReportSensorStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_orchestrator_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReportSensorStatusResponse.ProtoReflect.Descriptor instead.
func (*ReportSensorStatusResponse) Descriptor() ([]byte, []int) {
	return file_orchestrator_proto_rawDescGZIP(), []int{3}
}

func (x *ReportSensorStatusResponse) GetAck() string {
	if x != nil {
		return x.Ack
	}
	return ""
}

var File_orchestrator_proto protoreflect.FileDescriptor

var file_orchestrator_proto_rawDesc = []byte{
	0x0a, 0x12, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x22, 0x66, 0x0a,
	0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x73, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x1f, 0x0a,
	0x0b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x5f, 0x6d, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76,
	0x61, 0x6c, 0x4d, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x77, 0x6f, 0x72, 0x6b, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x77,
	0x6f, 0x72, 0x6b, 0x65, 0x72, 0x49, 0x64, 0x22, 0x2d, 0x0a, 0x14, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x22, 0x56, 0x0a, 0x19, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65, 0x6e,
	0x73, 0x6f, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x73, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x22, 0x33, 0x0a, 0x1a, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x15, 0x0a, 0x03, 0x61, 0x63,
	0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x61, 0x63, 0x6b, 0x32,
	0x6a, 0x0a, 0x0d, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x59, 0x0a, 0x0c, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x12, 0x23, 0x2e, 0x6f, 0x72, 0x63, 0x68,
	0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0x81, 0x01, 0x0a, 0x13, 0x4f, 0x72, 0x63,
	0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x6a, 0x0a, 0x12, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x29, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x53, 0x65, 0x6e,
	0x73, 0x6f, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x53, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x40, 0x5a,
	0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x73, 0x65, 0x71, 0x75, 0x69, 0x64,
	0x73, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f,
	0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62,
}

var (
	file_orchestrator_proto_rawDescOnce sync.Once
	file_orchestrator_proto_rawDescData = file_orchestrator_proto_rawDesc
)

func file_orchestrator_proto_rawDescGZIP() []byte {
	file_orchestrator_proto_rawDescOnce.Do(func() {
		file_orchestrator_proto_rawDescData = protoimpl.X.CompressGZIP(file_orchestrator_proto_rawDescData)
	})
	return file_orchestrator_proto_rawDescData
}

var file_orchestrator_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_orchestrator_proto_goTypes = []interface{}{
	(*CreateSensorRequest)(nil),
	(*CreateSensorResponse)(nil),
	(*ReportSensorStatusRequest)(nil),
	(*ReportSensorStatusResponse)(nil),
}
var file_orchestrator_proto_depIdxs = []int32{
	0, // 0: orchestratorpb.WorkerService.CreateSensor:input_type -> orchestratorpb.CreateSensorRequest
	2, // 1: orchestratorpb.OrchestratorService.ReportSensorStatus:input_type -> orchestratorpb.ReportSensorStatusRequest
	1, // 2: orchestratorpb.WorkerService.CreateSensor:output_type -> orchestratorpb.CreateSensorResponse
	3, // 3: orchestratorpb.OrchestratorService.ReportSensorStatus:output_type -> orchestratorpb.ReportSensorStatusResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_orchestrator_proto_init() }
func file_orchestrator_proto_init() {
	if File_orchestrator_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_orchestrator_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateSensorRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_orchestrator_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateSensorResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_orchestrator_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReportSensorStatusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_orchestrator_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReportSensorStatusResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_orchestrator_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_orchestrator_proto_goTypes,
		DependencyIndexes: file_orchestrator_proto_depIdxs,
		MessageInfos:      file_orchestrator_proto_msgTypes,
	}.Build()
	File_orchestrator_proto = out.File
	file_orchestrator_proto_rawDesc = nil
	file_orchestrator_proto_goTypes = nil
	file_orchestrator_proto_depIdxs = nil
}

var _ proto.Message = (*CreateSensorRequest)(nil)
var _ proto.Message = (*CreateSensorResponse)(nil)
var _ proto.Message = (*ReportSensorStatusRequest)(nil)
var _ proto.Message = (*ReportSensorStatusResponse)(nil)

