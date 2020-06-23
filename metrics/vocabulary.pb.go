// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.3
// source: metrics/vocabulary.proto

package gnostic_metrics_v1

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type WordCount struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Word  string `protobuf:"bytes,1,opt,name=word,proto3" json:"word,omitempty"`
	Count int32  `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *WordCount) Reset() {
	*x = WordCount{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metrics_vocabulary_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WordCount) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WordCount) ProtoMessage() {}

func (x *WordCount) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_vocabulary_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WordCount.ProtoReflect.Descriptor instead.
func (*WordCount) Descriptor() ([]byte, []int) {
	return file_metrics_vocabulary_proto_rawDescGZIP(), []int{0}
}

func (x *WordCount) GetWord() string {
	if x != nil {
		return x.Word
	}
	return ""
}

func (x *WordCount) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type Vocabulary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Schemas    []*WordCount `protobuf:"bytes,1,rep,name=schemas,proto3" json:"schemas,omitempty"`
	Properties []*WordCount `protobuf:"bytes,2,rep,name=properties,proto3" json:"properties,omitempty"`
	Operations []*WordCount `protobuf:"bytes,3,rep,name=operations,proto3" json:"operations,omitempty"`
	Parameters []*WordCount `protobuf:"bytes,4,rep,name=parameters,proto3" json:"parameters,omitempty"`
}

func (x *Vocabulary) Reset() {
	*x = Vocabulary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_metrics_vocabulary_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vocabulary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vocabulary) ProtoMessage() {}

func (x *Vocabulary) ProtoReflect() protoreflect.Message {
	mi := &file_metrics_vocabulary_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vocabulary.ProtoReflect.Descriptor instead.
func (*Vocabulary) Descriptor() ([]byte, []int) {
	return file_metrics_vocabulary_proto_rawDescGZIP(), []int{1}
}

func (x *Vocabulary) GetSchemas() []*WordCount {
	if x != nil {
		return x.Schemas
	}
	return nil
}

func (x *Vocabulary) GetProperties() []*WordCount {
	if x != nil {
		return x.Properties
	}
	return nil
}

func (x *Vocabulary) GetOperations() []*WordCount {
	if x != nil {
		return x.Operations
	}
	return nil
}

func (x *Vocabulary) GetParameters() []*WordCount {
	if x != nil {
		return x.Parameters
	}
	return nil
}

var File_metrics_vocabulary_proto protoreflect.FileDescriptor

var file_metrics_vocabulary_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f, 0x76, 0x6f, 0x63, 0x61, 0x62, 0x75,
	0x6c, 0x61, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x67, 0x6e, 0x6f, 0x73,
	0x74, 0x69, 0x63, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x76, 0x31, 0x22, 0x35,
	0x0a, 0x09, 0x57, 0x6f, 0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x77,
	0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x77, 0x6f, 0x72, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x82, 0x02, 0x0a, 0x0a, 0x56, 0x6f, 0x63, 0x61, 0x62, 0x75,
	0x6c, 0x61, 0x72, 0x79, 0x12, 0x37, 0x0a, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2e,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x64, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x52, 0x07, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x12, 0x3d, 0x0a,
	0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2e, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x3d, 0x0a, 0x0a,
	0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1d, 0x2e, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52,
	0x0a, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x3d, 0x0a, 0x0a, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1d, 0x2e, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x64, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x0a,
	0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x42, 0x1c, 0x5a, 0x1a, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x3b, 0x67, 0x6e, 0x6f, 0x73, 0x74, 0x69, 0x63, 0x5f, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_metrics_vocabulary_proto_rawDescOnce sync.Once
	file_metrics_vocabulary_proto_rawDescData = file_metrics_vocabulary_proto_rawDesc
)

func file_metrics_vocabulary_proto_rawDescGZIP() []byte {
	file_metrics_vocabulary_proto_rawDescOnce.Do(func() {
		file_metrics_vocabulary_proto_rawDescData = protoimpl.X.CompressGZIP(file_metrics_vocabulary_proto_rawDescData)
	})
	return file_metrics_vocabulary_proto_rawDescData
}

var file_metrics_vocabulary_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_metrics_vocabulary_proto_goTypes = []interface{}{
	(*WordCount)(nil),  // 0: gnostic.metrics.v1.WordCount
	(*Vocabulary)(nil), // 1: gnostic.metrics.v1.Vocabulary
}
var file_metrics_vocabulary_proto_depIdxs = []int32{
	0, // 0: gnostic.metrics.v1.Vocabulary.schemas:type_name -> gnostic.metrics.v1.WordCount
	0, // 1: gnostic.metrics.v1.Vocabulary.properties:type_name -> gnostic.metrics.v1.WordCount
	0, // 2: gnostic.metrics.v1.Vocabulary.operations:type_name -> gnostic.metrics.v1.WordCount
	0, // 3: gnostic.metrics.v1.Vocabulary.parameters:type_name -> gnostic.metrics.v1.WordCount
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_metrics_vocabulary_proto_init() }
func file_metrics_vocabulary_proto_init() {
	if File_metrics_vocabulary_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_metrics_vocabulary_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WordCount); i {
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
		file_metrics_vocabulary_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vocabulary); i {
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
			RawDescriptor: file_metrics_vocabulary_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_metrics_vocabulary_proto_goTypes,
		DependencyIndexes: file_metrics_vocabulary_proto_depIdxs,
		MessageInfos:      file_metrics_vocabulary_proto_msgTypes,
	}.Build()
	File_metrics_vocabulary_proto = out.File
	file_metrics_vocabulary_proto_rawDesc = nil
	file_metrics_vocabulary_proto_goTypes = nil
	file_metrics_vocabulary_proto_depIdxs = nil
}
