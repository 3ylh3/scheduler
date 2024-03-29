// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.4
// source: apiserver/qryjobinfo/query_job_info.proto

package qryjobinfo

import (
	job "github.com/3ylh3/scheduler/apiserver/job"
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

type QueryJobInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 任务id，-1代表查询所有
	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	// 任务名称
	JobName string `protobuf:"bytes,2,opt,name=job_name,json=jobName,proto3" json:"job_name,omitempty"`
}

func (x *QueryJobInfoRequest) Reset() {
	*x = QueryJobInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryJobInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryJobInfoRequest) ProtoMessage() {}

func (x *QueryJobInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryJobInfoRequest.ProtoReflect.Descriptor instead.
func (*QueryJobInfoRequest) Descriptor() ([]byte, []int) {
	return file_apiserver_qryjobinfo_query_job_info_proto_rawDescGZIP(), []int{0}
}

func (x *QueryJobInfoRequest) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

func (x *QueryJobInfoRequest) GetJobName() string {
	if x != nil {
		return x.JobName
	}
	return ""
}

type QueryJobInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result []*job.JobInfo `protobuf:"bytes,1,rep,name=result,proto3" json:"result,omitempty"`
}

func (x *QueryJobInfoResponse) Reset() {
	*x = QueryJobInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryJobInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryJobInfoResponse) ProtoMessage() {}

func (x *QueryJobInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryJobInfoResponse.ProtoReflect.Descriptor instead.
func (*QueryJobInfoResponse) Descriptor() ([]byte, []int) {
	return file_apiserver_qryjobinfo_query_job_info_proto_rawDescGZIP(), []int{1}
}

func (x *QueryJobInfoResponse) GetResult() []*job.JobInfo {
	if x != nil {
		return x.Result
	}
	return nil
}

var File_apiserver_qryjobinfo_query_job_info_proto protoreflect.FileDescriptor

var file_apiserver_qryjobinfo_query_job_info_proto_rawDesc = []byte{
	0x0a, 0x29, 0x61, 0x70, 0x69, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x71, 0x72, 0x79, 0x6a,
	0x6f, 0x62, 0x69, 0x6e, 0x66, 0x6f, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x5f, 0x6a, 0x6f, 0x62,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x71, 0x72, 0x79,
	0x6a, 0x6f, 0x62, 0x69, 0x6e, 0x66, 0x6f, 0x1a, 0x1c, 0x61, 0x70, 0x69, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x2f, 0x6a, 0x6f, 0x62, 0x2f, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x47, 0x0a, 0x13, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4a, 0x6f,
	0x62, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06,
	0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f,
	0x62, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x6a, 0x6f, 0x62, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6a, 0x6f, 0x62, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x3c,
	0x0a, 0x14, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4a, 0x6f, 0x62, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x6a, 0x6f, 0x62, 0x2e, 0x4a, 0x6f, 0x62,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0x63, 0x0a, 0x0c,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x4a, 0x6f, 0x62, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x53, 0x0a, 0x0c,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x4a, 0x6f, 0x62, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1f, 0x2e, 0x71,
	0x72, 0x79, 0x6a, 0x6f, 0x62, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4a,
	0x6f, 0x62, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e,
	0x71, 0x72, 0x79, 0x6a, 0x6f, 0x62, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x4a, 0x6f, 0x62, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x33, 0x79, 0x6c, 0x68, 0x33, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x72, 0x2f,
	0x61, 0x70, 0x69, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x71, 0x72, 0x79, 0x6a, 0x6f, 0x62,
	0x69, 0x6e, 0x66, 0x6f, 0x3b, 0x71, 0x72, 0x79, 0x6a, 0x6f, 0x62, 0x69, 0x6e, 0x66, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apiserver_qryjobinfo_query_job_info_proto_rawDescOnce sync.Once
	file_apiserver_qryjobinfo_query_job_info_proto_rawDescData = file_apiserver_qryjobinfo_query_job_info_proto_rawDesc
)

func file_apiserver_qryjobinfo_query_job_info_proto_rawDescGZIP() []byte {
	file_apiserver_qryjobinfo_query_job_info_proto_rawDescOnce.Do(func() {
		file_apiserver_qryjobinfo_query_job_info_proto_rawDescData = protoimpl.X.CompressGZIP(file_apiserver_qryjobinfo_query_job_info_proto_rawDescData)
	})
	return file_apiserver_qryjobinfo_query_job_info_proto_rawDescData
}

var file_apiserver_qryjobinfo_query_job_info_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_apiserver_qryjobinfo_query_job_info_proto_goTypes = []interface{}{
	(*QueryJobInfoRequest)(nil),  // 0: qryjobinfo.QueryJobInfoRequest
	(*QueryJobInfoResponse)(nil), // 1: qryjobinfo.QueryJobInfoResponse
	(*job.JobInfo)(nil),          // 2: job.JobInfo
}
var file_apiserver_qryjobinfo_query_job_info_proto_depIdxs = []int32{
	2, // 0: qryjobinfo.QueryJobInfoResponse.result:type_name -> job.JobInfo
	0, // 1: qryjobinfo.QueryJobInfo.QueryJobInfo:input_type -> qryjobinfo.QueryJobInfoRequest
	1, // 2: qryjobinfo.QueryJobInfo.QueryJobInfo:output_type -> qryjobinfo.QueryJobInfoResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_apiserver_qryjobinfo_query_job_info_proto_init() }
func file_apiserver_qryjobinfo_query_job_info_proto_init() {
	if File_apiserver_qryjobinfo_query_job_info_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryJobInfoRequest); i {
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
		file_apiserver_qryjobinfo_query_job_info_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryJobInfoResponse); i {
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
			RawDescriptor: file_apiserver_qryjobinfo_query_job_info_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_apiserver_qryjobinfo_query_job_info_proto_goTypes,
		DependencyIndexes: file_apiserver_qryjobinfo_query_job_info_proto_depIdxs,
		MessageInfos:      file_apiserver_qryjobinfo_query_job_info_proto_msgTypes,
	}.Build()
	File_apiserver_qryjobinfo_query_job_info_proto = out.File
	file_apiserver_qryjobinfo_query_job_info_proto_rawDesc = nil
	file_apiserver_qryjobinfo_query_job_info_proto_goTypes = nil
	file_apiserver_qryjobinfo_query_job_info_proto_depIdxs = nil
}
