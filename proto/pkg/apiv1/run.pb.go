// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// source: determined/api/v1/run.proto

package apiv1

import (
	_struct "github.com/golang/protobuf/ptypes/struct"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request to prepare to start reporting to a run.
type RunPrepareForReportingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// RunID to sync to.
	RunId int32 `protobuf:"varint,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	// Checkpoint storage config.
	CheckpointStorage *_struct.Struct `protobuf:"bytes,2,opt,name=checkpoint_storage,json=checkpointStorage,proto3,oneof" json:"checkpoint_storage,omitempty"`
}

func (x *RunPrepareForReportingRequest) Reset() {
	*x = RunPrepareForReportingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_run_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunPrepareForReportingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunPrepareForReportingRequest) ProtoMessage() {}

func (x *RunPrepareForReportingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_run_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunPrepareForReportingRequest.ProtoReflect.Descriptor instead.
func (*RunPrepareForReportingRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_run_proto_rawDescGZIP(), []int{0}
}

func (x *RunPrepareForReportingRequest) GetRunId() int32 {
	if x != nil {
		return x.RunId
	}
	return 0
}

func (x *RunPrepareForReportingRequest) GetCheckpointStorage() *_struct.Struct {
	if x != nil {
		return x.CheckpointStorage
	}
	return nil
}

// Response to prepare to start reporting to a run.
type RunPrepareForReportingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The storage_id to be used when creating new checkpoints. This will be
	// returned always when checkpoint storage is set in the request.
	StorageId *int32 `protobuf:"varint,1,opt,name=storage_id,json=storageId,proto3,oneof" json:"storage_id,omitempty"`
}

func (x *RunPrepareForReportingResponse) Reset() {
	*x = RunPrepareForReportingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_run_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunPrepareForReportingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunPrepareForReportingResponse) ProtoMessage() {}

func (x *RunPrepareForReportingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_run_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunPrepareForReportingResponse.ProtoReflect.Descriptor instead.
func (*RunPrepareForReportingResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_run_proto_rawDescGZIP(), []int{1}
}

func (x *RunPrepareForReportingResponse) GetStorageId() int32 {
	if x != nil && x.StorageId != nil {
		return *x.StorageId
	}
	return 0
}

var File_determined_api_v1_run_proto protoreflect.FileDescriptor

var file_determined_api_v1_run_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x72, 0x75, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x64,
	0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x73, 0x77, 0x61, 0x67, 0x67,
	0x65, 0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xaa, 0x01, 0x0a,
	0x1d, 0x52, 0x75, 0x6e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x46, 0x6f, 0x72, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15,
	0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x72, 0x75, 0x6e, 0x49, 0x64, 0x12, 0x4b, 0x0a, 0x12, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52, 0x11, 0x63, 0x68,
	0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x88,
	0x01, 0x01, 0x3a, 0x0e, 0x92, 0x41, 0x0b, 0x0a, 0x09, 0xd2, 0x01, 0x06, 0x72, 0x75, 0x6e, 0x5f,
	0x69, 0x64, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x22, 0x5a, 0x0a, 0x1e, 0x52, 0x75, 0x6e,
	0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x46, 0x6f, 0x72, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x0a, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x48,
	0x00, 0x52, 0x09, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x49, 0x64, 0x88, 0x01, 0x01, 0x3a,
	0x05, 0x92, 0x41, 0x02, 0x0a, 0x00, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x5f, 0x69, 0x64, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2d, 0x61,
	0x69, 0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_determined_api_v1_run_proto_rawDescOnce sync.Once
	file_determined_api_v1_run_proto_rawDescData = file_determined_api_v1_run_proto_rawDesc
)

func file_determined_api_v1_run_proto_rawDescGZIP() []byte {
	file_determined_api_v1_run_proto_rawDescOnce.Do(func() {
		file_determined_api_v1_run_proto_rawDescData = protoimpl.X.CompressGZIP(file_determined_api_v1_run_proto_rawDescData)
	})
	return file_determined_api_v1_run_proto_rawDescData
}

var file_determined_api_v1_run_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_determined_api_v1_run_proto_goTypes = []interface{}{
	(*RunPrepareForReportingRequest)(nil),  // 0: determined.api.v1.RunPrepareForReportingRequest
	(*RunPrepareForReportingResponse)(nil), // 1: determined.api.v1.RunPrepareForReportingResponse
	(*_struct.Struct)(nil),                 // 2: google.protobuf.Struct
}
var file_determined_api_v1_run_proto_depIdxs = []int32{
	2, // 0: determined.api.v1.RunPrepareForReportingRequest.checkpoint_storage:type_name -> google.protobuf.Struct
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_determined_api_v1_run_proto_init() }
func file_determined_api_v1_run_proto_init() {
	if File_determined_api_v1_run_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_determined_api_v1_run_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunPrepareForReportingRequest); i {
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
		file_determined_api_v1_run_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunPrepareForReportingResponse); i {
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
	file_determined_api_v1_run_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_determined_api_v1_run_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_determined_api_v1_run_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_determined_api_v1_run_proto_goTypes,
		DependencyIndexes: file_determined_api_v1_run_proto_depIdxs,
		MessageInfos:      file_determined_api_v1_run_proto_msgTypes,
	}.Build()
	File_determined_api_v1_run_proto = out.File
	file_determined_api_v1_run_proto_rawDesc = nil
	file_determined_api_v1_run_proto_goTypes = nil
	file_determined_api_v1_run_proto_depIdxs = nil
}
