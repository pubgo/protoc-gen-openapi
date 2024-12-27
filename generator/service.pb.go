// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.2
// source: openapiv3/service.proto

package generator

import (
	openapiv3 "github.com/google/gnostic-models/openapiv3"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Service struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tags                   []string                          `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty"`
	ExternalDocs           *openapiv3.ExternalDocs           `protobuf:"bytes,2,opt,name=external_docs,json=externalDocs,proto3" json:"external_docs,omitempty"`
	Parameters             []*openapiv3.ParameterOrReference `protobuf:"bytes,3,rep,name=parameters,proto3" json:"parameters,omitempty"`
	Security               []*openapiv3.SecurityRequirement  `protobuf:"bytes,4,rep,name=security,proto3" json:"security,omitempty"`
	Servers                []*openapiv3.Server               `protobuf:"bytes,5,rep,name=servers,proto3" json:"servers,omitempty"`
	SpecificationExtension []*openapiv3.NamedAny             `protobuf:"bytes,6,rep,name=specification_extension,json=specificationExtension,proto3" json:"specification_extension,omitempty"`
}

func (x *Service) Reset() {
	*x = Service{}
	mi := &file_openapiv3_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Service) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Service) ProtoMessage() {}

func (x *Service) ProtoReflect() protoreflect.Message {
	mi := &file_openapiv3_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Service.ProtoReflect.Descriptor instead.
func (*Service) Descriptor() ([]byte, []int) {
	return file_openapiv3_service_proto_rawDescGZIP(), []int{0}
}

func (x *Service) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Service) GetExternalDocs() *openapiv3.ExternalDocs {
	if x != nil {
		return x.ExternalDocs
	}
	return nil
}

func (x *Service) GetParameters() []*openapiv3.ParameterOrReference {
	if x != nil {
		return x.Parameters
	}
	return nil
}

func (x *Service) GetSecurity() []*openapiv3.SecurityRequirement {
	if x != nil {
		return x.Security
	}
	return nil
}

func (x *Service) GetServers() []*openapiv3.Server {
	if x != nil {
		return x.Servers
	}
	return nil
}

func (x *Service) GetSpecificationExtension() []*openapiv3.NamedAny {
	if x != nil {
		return x.SpecificationExtension
	}
	return nil
}

var file_openapiv3_service_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*Service)(nil),
		Field:         1144,
		Name:          "openapi.v3.service",
		Tag:           "bytes,1144,opt,name=service",
		Filename:      "openapiv3/service.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional openapi.v3.Service service = 1144;
	E_Service = &file_openapiv3_service_proto_extTypes[0]
)

var File_openapiv3_service_proto protoreflect.FileDescriptor

var file_openapiv3_service_proto_rawDesc = []byte{
	0x0a, 0x17, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x33, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x6f, 0x70, 0x65, 0x6e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x33, 0x1a, 0x19, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x33,
	0x2f, 0x4f, 0x70, 0x65, 0x6e, 0x41, 0x50, 0x49, 0x76, 0x33, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xd8, 0x02, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x12, 0x3d, 0x0a, 0x0d, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x5f, 0x64,
	0x6f, 0x63, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x33, 0x2e, 0x45, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x44,
	0x6f, 0x63, 0x73, 0x52, 0x0c, 0x65, 0x78, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x44, 0x6f, 0x63,
	0x73, 0x12, 0x40, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x4f, 0x72, 0x52, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74,
	0x65, 0x72, 0x73, 0x12, 0x3b, 0x0a, 0x08, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x33, 0x2e, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x69,
	0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79,
	0x12, 0x2c, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x33, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x12, 0x4d,
	0x0a, 0x17, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x33, 0x2e, 0x4e, 0x61, 0x6d,
	0x65, 0x64, 0x41, 0x6e, 0x79, 0x52, 0x16, 0x73, 0x70, 0x65, 0x63, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x3a, 0x4f, 0x0a,
	0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xf8, 0x08, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x33, 0x2e, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x17,
	0x5a, 0x15, 0x2e, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x3b, 0x67, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_openapiv3_service_proto_rawDescOnce sync.Once
	file_openapiv3_service_proto_rawDescData = file_openapiv3_service_proto_rawDesc
)

func file_openapiv3_service_proto_rawDescGZIP() []byte {
	file_openapiv3_service_proto_rawDescOnce.Do(func() {
		file_openapiv3_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_openapiv3_service_proto_rawDescData)
	})
	return file_openapiv3_service_proto_rawDescData
}

var file_openapiv3_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_openapiv3_service_proto_goTypes = []any{
	(*Service)(nil),                        // 0: openapi.v3.Service
	(*openapiv3.ExternalDocs)(nil),         // 1: openapi.v3.ExternalDocs
	(*openapiv3.ParameterOrReference)(nil), // 2: openapi.v3.ParameterOrReference
	(*openapiv3.SecurityRequirement)(nil),  // 3: openapi.v3.SecurityRequirement
	(*openapiv3.Server)(nil),               // 4: openapi.v3.Server
	(*openapiv3.NamedAny)(nil),             // 5: openapi.v3.NamedAny
	(*descriptorpb.ServiceOptions)(nil),    // 6: google.protobuf.ServiceOptions
}
var file_openapiv3_service_proto_depIdxs = []int32{
	1, // 0: openapi.v3.Service.external_docs:type_name -> openapi.v3.ExternalDocs
	2, // 1: openapi.v3.Service.parameters:type_name -> openapi.v3.ParameterOrReference
	3, // 2: openapi.v3.Service.security:type_name -> openapi.v3.SecurityRequirement
	4, // 3: openapi.v3.Service.servers:type_name -> openapi.v3.Server
	5, // 4: openapi.v3.Service.specification_extension:type_name -> openapi.v3.NamedAny
	6, // 5: openapi.v3.service:extendee -> google.protobuf.ServiceOptions
	0, // 6: openapi.v3.service:type_name -> openapi.v3.Service
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	6, // [6:7] is the sub-list for extension type_name
	5, // [5:6] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_openapiv3_service_proto_init() }
func file_openapiv3_service_proto_init() {
	if File_openapiv3_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_openapiv3_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_openapiv3_service_proto_goTypes,
		DependencyIndexes: file_openapiv3_service_proto_depIdxs,
		MessageInfos:      file_openapiv3_service_proto_msgTypes,
		ExtensionInfos:    file_openapiv3_service_proto_extTypes,
	}.Build()
	File_openapiv3_service_proto = out.File
	file_openapiv3_service_proto_rawDesc = nil
	file_openapiv3_service_proto_goTypes = nil
	file_openapiv3_service_proto_depIdxs = nil
}
