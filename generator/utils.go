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
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// contains returns true if an array contains a specified string.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// appendUnique appends a string, to a string slice, if the string is not already in the slice
func appendUnique(s []string, e string) []string {
	if !contains(s, e) {
		return append(s, e)
	}
	return s
}

// singular produces the singular form of a collection name.
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

// formatMessageName 格式化消息名称
func formatMessageName(message protoreflect.MessageDescriptor) string {
	typeName := fullMessageTypeName(message)
	name := getMessageName(message)
	if typeName == ".google.protobuf.Value" {
		name = protobufValueName
	} else if typeName == ".google.protobuf.Any" {
		name = protobufAnyName
	}

	if len(name) > 1 {
		name = strings.ToUpper(name[0:1]) + name[1:]
	}

	if len(name) == 1 {
		name = strings.ToLower(name)
	}

	return name
}

// formatFieldName 格式化字段名称
func formatFieldName(field protoreflect.FieldDescriptor) string {
	return field.JSONName()
}

// fullMessageTypeName 获取完整的消息类型名称
func fullMessageTypeName(message protoreflect.MessageDescriptor) string {
	return string(message.FullName())
}

// getMessageName 获取消息名称
func getMessageName(message protoreflect.MessageDescriptor) string {
	prefix := ""
	parent := message.Parent()
	if _, ok := parent.(protoreflect.MessageDescriptor); ok {
		prefix = string(parent.Name()) + "_" + prefix
	}
	return prefix + string(message.Name())
}

// getValueKind 获取值类型
func getValueKind(message protoreflect.MessageDescriptor) string {
	valueField := getValueField(message)
	return valueField.Kind().String()
}

// getValueField 获取值字段
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

// schemaReferenceForMessage 获取消息的 Schema 引用
func schemaReferenceForMessage(message protoreflect.MessageDescriptor) string {
	schemaName := formatMessageName(message)
	return "#/components/schemas/" + schemaName
}

// filterCommentString 过滤注释字符串
func filterCommentString(comment protogen.Comments) string {
	// TODO: 实现注释过滤逻辑
	return string(comment)
}
