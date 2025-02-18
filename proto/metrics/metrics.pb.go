// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v5.29.2
// source: proto/metrics/metrics.proto

package metrics

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Metric struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	MName         string                 `protobuf:"bytes,2,opt,name=m_name,json=mName,proto3" json:"m_name,omitempty"`
	Value         *float64               `protobuf:"fixed64,3,opt,name=value,proto3,oneof" json:"value,omitempty"`
	Delta         *int64                 `protobuf:"varint,4,opt,name=delta,proto3,oneof" json:"delta,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Metric) Reset() {
	*x = Metric{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{0}
}

func (x *Metric) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metric) GetMName() string {
	if x != nil {
		return x.MName
	}
	return ""
}

func (x *Metric) GetValue() float64 {
	if x != nil && x.Value != nil {
		return *x.Value
	}
	return 0
}

func (x *Metric) GetDelta() int64 {
	if x != nil && x.Delta != nil {
		return *x.Delta
	}
	return 0
}

type CreateMetricRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metric        *Metric                `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateMetricRequest) Reset() {
	*x = CreateMetricRequest{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateMetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateMetricRequest) ProtoMessage() {}

func (x *CreateMetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateMetricRequest.ProtoReflect.Descriptor instead.
func (*CreateMetricRequest) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{1}
}

func (x *CreateMetricRequest) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type CreateMetricResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateMetricResponse) Reset() {
	*x = CreateMetricResponse{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateMetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateMetricResponse) ProtoMessage() {}

func (x *CreateMetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateMetricResponse.ProtoReflect.Descriptor instead.
func (*CreateMetricResponse) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{2}
}

func (x *CreateMetricResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type CreateMetricsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metrics       []*Metric              `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateMetricsRequest) Reset() {
	*x = CreateMetricsRequest{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateMetricsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateMetricsRequest) ProtoMessage() {}

func (x *CreateMetricsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateMetricsRequest.ProtoReflect.Descriptor instead.
func (*CreateMetricsRequest) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{3}
}

func (x *CreateMetricsRequest) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type CreateMetricsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateMetricsResponse) Reset() {
	*x = CreateMetricsResponse{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateMetricsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateMetricsResponse) ProtoMessage() {}

func (x *CreateMetricsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateMetricsResponse.ProtoReflect.Descriptor instead.
func (*CreateMetricsResponse) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{4}
}

func (x *CreateMetricsResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type GetMetricRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"` // Using ID directly to fetch a specific metric
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMetricRequest) Reset() {
	*x = GetMetricRequest{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetricRequest) ProtoMessage() {}

func (x *GetMetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetricRequest.ProtoReflect.Descriptor instead.
func (*GetMetricRequest) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{5}
}

func (x *GetMetricRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetMetricResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metric        *Metric                `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetMetricResponse) Reset() {
	*x = GetMetricResponse{}
	mi := &file_proto_metrics_metrics_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetMetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetricResponse) ProtoMessage() {}

func (x *GetMetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_metrics_metrics_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetricResponse.ProtoReflect.Descriptor instead.
func (*GetMetricResponse) Descriptor() ([]byte, []int) {
	return file_proto_metrics_metrics_proto_rawDescGZIP(), []int{6}
}

func (x *GetMetricResponse) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

func (x *GetMetricResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_proto_metrics_metrics_proto protoreflect.FileDescriptor

var file_proto_metrics_metrics_proto_rawDesc = string([]byte{
	0x0a, 0x1b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x79, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x15,
	0x0a, 0x06, 0x6d, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6d, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x01, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x88, 0x01, 0x01,
	0x12, 0x19, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x48,
	0x01, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x22,
	0x3e, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73,
	0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x22,
	0x30, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x41, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x07, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x07, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x22, 0x31, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x22, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x56, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x27, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x32, 0xba, 0x02, 0x0a, 0x07, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12,
	0x63, 0x0a, 0x0c, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12,
	0x1c, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x10, 0x3a, 0x01, 0x2a, 0x22, 0x0b, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x12, 0x6c, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x1d, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x3a, 0x01, 0x2a, 0x22,
	0x11, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x61, 0x74,
	0x63, 0x68, 0x12, 0x5c, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12,
	0x19, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10,
	0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d,
	0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d,
	0x69, 0x68, 0x61, 0x69, 0x6c, 0x74, 0x75, 0x64, 0x6f, 0x73, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_proto_metrics_metrics_proto_rawDescOnce sync.Once
	file_proto_metrics_metrics_proto_rawDescData []byte
)

func file_proto_metrics_metrics_proto_rawDescGZIP() []byte {
	file_proto_metrics_metrics_proto_rawDescOnce.Do(func() {
		file_proto_metrics_metrics_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_metrics_metrics_proto_rawDesc), len(file_proto_metrics_metrics_proto_rawDesc)))
	})
	return file_proto_metrics_metrics_proto_rawDescData
}

var file_proto_metrics_metrics_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_proto_metrics_metrics_proto_goTypes = []any{
	(*Metric)(nil),                // 0: metrics.Metric
	(*CreateMetricRequest)(nil),   // 1: metrics.CreateMetricRequest
	(*CreateMetricResponse)(nil),  // 2: metrics.CreateMetricResponse
	(*CreateMetricsRequest)(nil),  // 3: metrics.CreateMetricsRequest
	(*CreateMetricsResponse)(nil), // 4: metrics.CreateMetricsResponse
	(*GetMetricRequest)(nil),      // 5: metrics.GetMetricRequest
	(*GetMetricResponse)(nil),     // 6: metrics.GetMetricResponse
}
var file_proto_metrics_metrics_proto_depIdxs = []int32{
	0, // 0: metrics.CreateMetricRequest.metric:type_name -> metrics.Metric
	0, // 1: metrics.CreateMetricsRequest.metrics:type_name -> metrics.Metric
	0, // 2: metrics.GetMetricResponse.metric:type_name -> metrics.Metric
	1, // 3: metrics.Metrics.CreateMetric:input_type -> metrics.CreateMetricRequest
	3, // 4: metrics.Metrics.CreateMetrics:input_type -> metrics.CreateMetricsRequest
	5, // 5: metrics.Metrics.GetMetric:input_type -> metrics.GetMetricRequest
	2, // 6: metrics.Metrics.CreateMetric:output_type -> metrics.CreateMetricResponse
	4, // 7: metrics.Metrics.CreateMetrics:output_type -> metrics.CreateMetricsResponse
	6, // 8: metrics.Metrics.GetMetric:output_type -> metrics.GetMetricResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_metrics_metrics_proto_init() }
func file_proto_metrics_metrics_proto_init() {
	if File_proto_metrics_metrics_proto != nil {
		return
	}
	file_proto_metrics_metrics_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_metrics_metrics_proto_rawDesc), len(file_proto_metrics_metrics_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_metrics_metrics_proto_goTypes,
		DependencyIndexes: file_proto_metrics_metrics_proto_depIdxs,
		MessageInfos:      file_proto_metrics_metrics_proto_msgTypes,
	}.Build()
	File_proto_metrics_metrics_proto = out.File
	file_proto_metrics_metrics_proto_goTypes = nil
	file_proto_metrics_metrics_proto_depIdxs = nil
}
