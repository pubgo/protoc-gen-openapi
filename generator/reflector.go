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
//

package generator

import (
	"log"
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
)

const (
	protobufValueName = "GoogleProtobufValue"
	protobufAnyName   = "GoogleProtobufAny"
)

// OpenAPIv3Reflector 用于将 Protocol Buffer 类型转换为 OpenAPI 类型的反射器
type OpenAPIv3Reflector struct {
	conf Configuration // 生成器配置

	requiredSchemas []string // 通过引用使用的模式的名称
}

// NewOpenAPIv3Reflector 创建一个新的反射器
// 参数:
//   - conf: 生成器配置
//
// 返回值:
//   - 创建的 OpenAPIv3Reflector 实例
func NewOpenAPIv3Reflector(conf Configuration) *OpenAPIv3Reflector {
	cfg := DefaultConfig()
	cfg = cfg.Merge(conf)

	return &OpenAPIv3Reflector{
		conf:            cfg,
		requiredSchemas: make([]string, 0),
	}
}

// formatMessageName 格式化消息名称
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 格式化后的消息名称
func (r *OpenAPIv3Reflector) formatMessageName(message protoreflect.MessageDescriptor) string {
	typeName := fullMessageTypeName(message)
	name := getMessageName(message)

	// 使用特殊名称处理内置类型
	if !*r.conf.FQSchemaNaming {
		name = r.handleBuiltInTypes(typeName, name)
	}

	// 根据命名配置处理名称格式
	name = r.applyNamingConventions(name)

	// 是否使用完全限定名称
	if *r.conf.FQSchemaNaming {
		packageName := string(message.ParentFile().Package())
		name = packageName + "." + name
	}

	return name
}

// handleBuiltInTypes 处理内置类型的命名
// 参数:
//   - typeName: 类型全名
//   - name: 原始名称
//
// 返回值:
//   - 处理后的名称
func (r *OpenAPIv3Reflector) handleBuiltInTypes(typeName, name string) string {
	if typeName == ".google.protobuf.Value" {
		return protobufValueName
	} else if typeName == ".google.protobuf.Any" {
		return protobufAnyName
	}
	return name
}

// applyNamingConventions 应用命名约定
// 参数:
//   - name: 原始名称
//
// 返回值:
//   - 应用命名约定后的名称
func (r *OpenAPIv3Reflector) applyNamingConventions(name string) string {
	if *r.conf.Naming == "json" {
		if len(name) > 1 {
			name = strings.ToUpper(name[0:1]) + name[1:]
		}

		if len(name) == 1 {
			name = strings.ToLower(name)
		}
	}
	return name
}

// formatFieldName 格式化字段名称
// 参数:
//   - field: 字段描述符
//
// 返回值:
//   - 格式化后的字段名称
func (r *OpenAPIv3Reflector) formatFieldName(field protoreflect.FieldDescriptor) string {
	if *r.conf.Naming == "proto" {
		return string(field.Name())
	}

	return field.JSONName()
}

// responseContentForMessage 为消息生成响应内容
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 状态码
//   - 媒体类型
func (r *OpenAPIv3Reflector) responseContentForMessage(message protoreflect.MessageDescriptor) (string, *v3.MediaTypes) {
	typeName := fullMessageTypeName(message)

	// 处理特殊类型
	if r.isEmptyMessage(typeName) {
		return "200", &v3.MediaTypes{}
	}

	if r.isHttpBodyMessage(typeName) {
		return "200", wellknown.NewGoogleApiHttpBodyMediaType()
	}

	// 处理普通消息类型
	return "200", wellknown.NewApplicationJsonMediaType(r.schemaOrReferenceForMessage(message))
}

// isEmptyMessage 检查消息是否为 Empty 类型
// 参数:
//   - typeName: 类型全名
//
// 返回值:
//   - 是否为 Empty 类型
func (r *OpenAPIv3Reflector) isEmptyMessage(typeName string) bool {
	return typeName == ".google.protobuf.Empty"
}

// isHttpBodyMessage 检查消息是否为 HttpBody 类型
// 参数:
//   - typeName: 类型全名
//
// 返回值:
//   - 是否为 HttpBody 类型
func (r *OpenAPIv3Reflector) isHttpBodyMessage(typeName string) bool {
	return typeName == ".google.api.HttpBody"
}

// schemaReferenceForMessage 获取消息的模式引用
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 模式引用
func (r *OpenAPIv3Reflector) schemaReferenceForMessage(message protoreflect.MessageDescriptor) string {
	schemaName := r.formatMessageName(message)
	if !contains(r.requiredSchemas, schemaName) {
		r.requiredSchemas = append(r.requiredSchemas, schemaName)
	}
	return "#/components/schemas/" + schemaName
}

// schemaOrReferenceForMessage 获取消息的模式或引用
// 对于简单类型，返回完整模式；对于复杂类型，返回引用到 `#/components/schemas/` 中的定义
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 模式或引用
func (r *OpenAPIv3Reflector) schemaOrReferenceForMessage(message protoreflect.MessageDescriptor) *v3.SchemaOrReference {
	typeName := fullMessageTypeName(message)

	// 处理特殊类型
	if schema := r.handleSpecialMessageTypes(typeName, message); schema != nil {
		return schema
	}

	// 处理普通类型
	ref := r.schemaReferenceForMessage(message)
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Reference{Reference: &v3.Reference{XRef: ref}},
	}
}

// handleSpecialMessageTypes 处理特殊消息类型
// 参数:
//   - typeName: 类型全名
//   - message: 消息描述符
//
// 返回值:
//   - 特殊类型的模式，如果不是特殊类型则返回 nil
func (r *OpenAPIv3Reflector) handleSpecialMessageTypes(typeName string, message protoreflect.MessageDescriptor) *v3.SchemaOrReference {
	switch typeName {
	case ".google.api.HttpBody":
		return wellknown.NewGoogleApiHttpBodySchema()

	case ".google.protobuf.Timestamp":
		return wellknown.NewGoogleProtobufTimestampSchema()

	case ".google.protobuf.Duration":
		return wellknown.NewGoogleProtobufDurationSchema()

	case ".google.type.Date":
		return wellknown.NewGoogleTypeDateSchema()

	case ".google.type.DateTime":
		return wellknown.NewGoogleTypeDateTimeSchema()

	case ".google.protobuf.FieldMask":
		return wellknown.NewGoogleProtobufFieldMaskSchema()

	case ".google.protobuf.Struct":
		return wellknown.NewGoogleProtobufStructSchema()

	case ".google.protobuf.Empty":
		// Empty 更接近 JSON undefined 而非 null，因此忽略此字段
		return nil

	case ".google.protobuf.BoolValue":
		return wellknown.NewBooleanSchema()

	case ".google.protobuf.BytesValue":
		return wellknown.NewBytesSchema()

	case ".google.protobuf.Int32Value", ".google.protobuf.UInt32Value":
		return wellknown.NewIntegerSchema(getValueKind(message))

	case ".google.protobuf.StringValue", ".google.protobuf.Int64Value", ".google.protobuf.UInt64Value":
		return wellknown.NewStringSchema()

	case ".google.protobuf.FloatValue", ".google.protobuf.DoubleValue":
		return wellknown.NewNumberSchema(getValueKind(message))

	default:
		return nil
	}
}

// schemaOrReferenceForField 获取字段的模式或引用
// 参数:
//   - field: 字段
//   - desc: 字段描述符，可选，如果为 nil 则使用 field.Desc
//
// 返回值:
//   - 模式或引用
func (r *OpenAPIv3Reflector) schemaOrReferenceForField(field *protogen.Field, desc protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	if desc == nil {
		desc = field.Desc
	}

	var kindSchema *v3.SchemaOrReference
	kind := desc.Kind()

	// 处理不同类型的字段
	kindSchema = r.handleFieldByKind(field, desc, kind)

	// 处理列表类型
	if field.Desc.IsList() {
		kindSchema = wellknown.NewListSchema(kindSchema)
	}

	return kindSchema
}

// handleFieldByKind 根据字段类型处理字段
// 参数:
//   - field: 字段
//   - desc: 字段描述符
//   - kind: 字段类型
//
// 返回值:
//   - 模式或引用
func (r *OpenAPIv3Reflector) handleFieldByKind(field *protogen.Field, desc protoreflect.FieldDescriptor, kind protoreflect.Kind) *v3.SchemaOrReference {
	switch kind {
	case protoreflect.MessageKind:
		return r.handleMessageField(field, desc)
	case protoreflect.StringKind:
		return wellknown.NewStringSchema()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		return wellknown.NewIntegerSchema(kind.String())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		return wellknown.NewStringSchema()
	case protoreflect.EnumKind:
		return wellknown.NewEnumSchema(r.conf.EnumType, field)
	case protoreflect.BoolKind:
		return wellknown.NewBooleanSchema()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return wellknown.NewNumberSchema(kind.String())
	case protoreflect.BytesKind:
		return wellknown.NewBytesSchema()
	default:
		log.Printf("(TODO) Unsupported field type: %+v", fullMessageTypeName(field.Desc.Message()))
		return nil
	}
}

// handleMessageField 处理消息类型字段
// 参数:
//   - field: 字段
//   - desc: 字段描述符
//
// 返回值:
//   - 模式或引用
func (r *OpenAPIv3Reflector) handleMessageField(field *protogen.Field, desc protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	if desc.IsMap() {
		// 处理映射类型
		return wellknown.NewGoogleProtobufMapFieldEntrySchema(r.schemaOrReferenceForField(field, desc.MapValue()))
	} else {
		// 处理普通消息类型
		return r.schemaOrReferenceForMessage(desc.Message())
	}
}
