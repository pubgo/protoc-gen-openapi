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
	"unicode"

	v3 "github.com/google/gnostic-models/openapiv3"
	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// contains 检查字符串切片是否包含指定字符串
// 参数:
//   - s: 要检查的字符串切片
//   - e: 要查找的字符串
//
// 返回值:
//   - 如果s中包含e则返回true，否则返回false
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// appendUnique 将字符串添加到切片中（如果切片中不存在该字符串）
// 参数:
//   - s: 目标字符串切片
//   - e: 要添加的字符串
//
// 返回值:
//   - 添加后的字符串切片
func appendUnique(s []string, e string) []string {
	if !contains(s, e) {
		return append(s, e)
	}
	return s
}

// singular 将复数形式的单词转换为单数形式
// 参数:
//   - plural: 复数形式的单词
//
// 返回值:
//   - 单数形式的单词
func singular(plural string) string {
	if strings.HasSuffix(plural, "ves") {
		return strings.TrimSuffix(plural, "ves") + "f"
	}
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}

// toUpperFirstLetter 将字符串的第一个字母转换为大写
// 参数:
//   - s: 要转换的字符串
//
// 返回值:
//   - 转换后的字符串
func toUpperFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// toLowerFirstLetter 将字符串的第一个字母转换为小写
// 参数:
//   - s: 要转换的字符串
//
// 返回值:
//   - 转换后的字符串
func toLowerFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// fullMessageTypeName 获取完整的消息类型名称
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 完整的消息类型名称
func fullMessageTypeName(message protoreflect.MessageDescriptor) string {
	return string(message.FullName())
}

// getMessageName 获取消息名称
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 消息名称
func getMessageName(message protoreflect.MessageDescriptor) string {
	prefix := ""
	parent := message.Parent()
	if _, ok := parent.(protoreflect.MessageDescriptor); ok {
		prefix = string(parent.Name()) + "_" + prefix
	}
	return prefix + string(message.Name())
}

// getValueKind 获取值类型
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 值类型的字符串表示
func getValueKind(message protoreflect.MessageDescriptor) string {
	valueField := getValueField(message)
	return valueField.Kind().String()
}

// getValueField 获取值字段
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 值字段的描述符
func getValueField(message protoreflect.MessageDescriptor) protoreflect.FieldDescriptor {
	fields := message.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.Name() == "value" {
			return field
		}
	}
	return nil
}

// applyNamingConventions 应用命名约定
// 参数:
//   - name: 原始名称
//   - naming: 命名约定类型 ("proto" 或 "json")
//
// 返回值:
//   - 应用命名约定后的名称
func applyNamingConventions(name, naming string) string {
	if naming == "json" {
		if len(name) > 1 {
			name = toUpperFirstLetter(name)
		}

		if len(name) == 1 {
			name = strings.ToLower(name)
		}
	}
	return name
}

// isEmptyMessage 检查消息是否为 Empty 类型
// 参数:
//   - typeName: 类型全名
//
// 返回值:
//   - 是否为 Empty 类型
func isEmptyMessage(typeName string) bool {
	return typeName == ".google.protobuf.Empty"
}

// isHttpBodyMessage 检查消息是否为 HttpBody 类型
// 参数:
//   - typeName: 类型全名
//
// 返回值:
//   - 是否为 HttpBody 类型
func isHttpBodyMessage(typeName string) bool {
	return typeName == ".google.api.HttpBody"
}

// handleSpecialMessageTypes 处理特殊消息类型
// 参数:
//   - typeName: 类型全名
//   - message: 消息描述符
//
// 返回值:
//   - 特殊类型的模式，如果不是特殊类型则返回 nil
func handleSpecialMessageTypes(typeName string, message protoreflect.MessageDescriptor) *v3.SchemaOrReference {
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

// handleFieldByKind 根据字段类型处理字段
// 参数:
//   - field: 字段
//   - desc: 字段描述符
//   - kind: 字段类型
//   - enumType: 枚举类型
//   - schemaOrReferenceForField: 用于获取字段模式的函数
//   - schemaOrReferenceForMessage: 用于获取消息模式的函数
//
// 返回值:
//   - 模式或引用
func handleFieldByKind(
	field *protogen.Field,
	desc protoreflect.FieldDescriptor,
	kind protoreflect.Kind,
	enumType string,
	schemaOrReferenceForField func(*protogen.Field, protoreflect.FieldDescriptor) *v3.SchemaOrReference,
	schemaOrReferenceForMessage func(protoreflect.MessageDescriptor) *v3.SchemaOrReference,
) *v3.SchemaOrReference {
	switch kind {
	case protoreflect.MessageKind:
		if desc.IsMap() {
			// 处理映射类型
			return wellknown.NewGoogleProtobufMapFieldEntrySchema(schemaOrReferenceForField(field, desc.MapValue()))
		} else {
			// 处理普通消息类型
			return schemaOrReferenceForMessage(desc.Message())
		}
	case protoreflect.StringKind:
		return wellknown.NewStringSchema()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		return wellknown.NewIntegerSchema(kind.String())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		return wellknown.NewStringSchema()
	case protoreflect.EnumKind:
		return wellknown.NewEnumSchema(&enumType, field)
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
//   - schemaOrReferenceForField: 用于获取字段模式的函数
//   - schemaOrReferenceForMessage: 用于获取消息模式的函数
//
// 返回值:
//   - 模式或引用

// initializeDocument 初始化 OpenAPI 文档
// 参数:
//   - conf: 生成器配置
//
// 返回值:
//   - 初始化后的 OpenAPI 文档
func initializeDocument(conf Configuration) *v3.Document {
	// 创建 OpenAPI 文档
	document := &v3.Document{
		Openapi: "3.0.0",
		Info: &v3.Info{
			Title:       *conf.Title,
			Version:     *conf.Version,
			Description: *conf.Description,
		},
		Components: &v3.Components{
			Schemas: &v3.SchemasOrReferences{
				AdditionalProperties: []*v3.NamedSchemaOrReference{},
			},
		},
	}

	return document
}
