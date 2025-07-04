// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v4.25.3
// source: metric_update.proto

package protos

import (
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
	Type          string                 `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Value         float64                `protobuf:"fixed64,3,opt,name=value,proto3" json:"value,omitempty"`
	Delta         int64                  `protobuf:"varint,4,opt,name=delta,proto3" json:"delta,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Metric) Reset() {
	*x = Metric{}
	mi := &file_metric_update_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_metric_update_proto_msgTypes[0]
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
	return file_metric_update_proto_rawDescGZIP(), []int{0}
}

func (x *Metric) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metric) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Metric) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *Metric) GetDelta() int64 {
	if x != nil {
		return x.Delta
	}
	return 0
}

type UpdateMetricsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metrics       []*Metric              `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateMetricsRequest) Reset() {
	*x = UpdateMetricsRequest{}
	mi := &file_metric_update_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateMetricsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateMetricsRequest) ProtoMessage() {}

func (x *UpdateMetricsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_metric_update_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateMetricsRequest.ProtoReflect.Descriptor instead.
func (*UpdateMetricsRequest) Descriptor() ([]byte, []int) {
	return file_metric_update_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateMetricsRequest) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

type UpdateMetricsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Metrics       []*Metric              `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
	Error         string                 `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateMetricsResponse) Reset() {
	*x = UpdateMetricsResponse{}
	mi := &file_metric_update_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateMetricsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateMetricsResponse) ProtoMessage() {}

func (x *UpdateMetricsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_metric_update_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateMetricsResponse.ProtoReflect.Descriptor instead.
func (*UpdateMetricsResponse) Descriptor() ([]byte, []int) {
	return file_metric_update_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateMetricsResponse) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

func (x *UpdateMetricsResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_metric_update_proto protoreflect.FileDescriptor

const file_metric_update_proto_rawDesc = "" +
	"\n" +
	"\x13metric_update.proto\x12\x13go_yandex_practicum\"X\n" +
	"\x06Metric\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04type\x18\x02 \x01(\tR\x04type\x12\x14\n" +
	"\x05value\x18\x03 \x01(\x01R\x05value\x12\x14\n" +
	"\x05delta\x18\x04 \x01(\x03R\x05delta\"M\n" +
	"\x14UpdateMetricsRequest\x125\n" +
	"\ametrics\x18\x01 \x03(\v2\x1b.go_yandex_practicum.MetricR\ametrics\"d\n" +
	"\x15UpdateMetricsResponse\x125\n" +
	"\ametrics\x18\x01 \x03(\v2\x1b.go_yandex_practicum.MetricR\ametrics\x12\x14\n" +
	"\x05error\x18\x02 \x01(\tR\x05error2q\n" +
	"\rMetricUpdater\x12`\n" +
	"\aUpdates\x12).go_yandex_practicum.UpdateMetricsRequest\x1a*.go_yandex_practicum.UpdateMetricsResponseB4Z2github.com/sbilibin2017/go-yandex-practicum/protosb\x06proto3"

var (
	file_metric_update_proto_rawDescOnce sync.Once
	file_metric_update_proto_rawDescData []byte
)

func file_metric_update_proto_rawDescGZIP() []byte {
	file_metric_update_proto_rawDescOnce.Do(func() {
		file_metric_update_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_metric_update_proto_rawDesc), len(file_metric_update_proto_rawDesc)))
	})
	return file_metric_update_proto_rawDescData
}

var file_metric_update_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_metric_update_proto_goTypes = []any{
	(*Metric)(nil),                // 0: go_yandex_practicum.Metric
	(*UpdateMetricsRequest)(nil),  // 1: go_yandex_practicum.UpdateMetricsRequest
	(*UpdateMetricsResponse)(nil), // 2: go_yandex_practicum.UpdateMetricsResponse
}
var file_metric_update_proto_depIdxs = []int32{
	0, // 0: go_yandex_practicum.UpdateMetricsRequest.metrics:type_name -> go_yandex_practicum.Metric
	0, // 1: go_yandex_practicum.UpdateMetricsResponse.metrics:type_name -> go_yandex_practicum.Metric
	1, // 2: go_yandex_practicum.MetricUpdater.Updates:input_type -> go_yandex_practicum.UpdateMetricsRequest
	2, // 3: go_yandex_practicum.MetricUpdater.Updates:output_type -> go_yandex_practicum.UpdateMetricsResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_metric_update_proto_init() }
func file_metric_update_proto_init() {
	if File_metric_update_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_metric_update_proto_rawDesc), len(file_metric_update_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_metric_update_proto_goTypes,
		DependencyIndexes: file_metric_update_proto_depIdxs,
		MessageInfos:      file_metric_update_proto_msgTypes,
	}.Build()
	File_metric_update_proto = out.File
	file_metric_update_proto_goTypes = nil
	file_metric_update_proto_depIdxs = nil
}
