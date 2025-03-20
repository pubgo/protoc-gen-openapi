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
	"unicode"

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

// formatMessageName 格式化消息名称
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 格式化后的消息名称
func formatMessageName(message protoreflect.MessageDescriptor) string {
	typeName := fullMessageTypeName(message)
	name := getMessageName(message)
	if typeName == ".google.protobuf.Value" {
		name = protobufValueName
	} else if typeName == ".google.protobuf.Any" {
		name = protobufAnyName
	}

	// 首字母大写，其余保持不变
	if len(name) > 1 {
		name = toUpperFirstLetter(name)
	}

	// 对于单个字母的名称，全部小写
	if len(name) == 1 {
		name = strings.ToLower(name)
	}

	return name
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

// formatFieldName 格式化字段名称
// 参数:
//   - field: 字段描述符
//
// 返回值:
//   - 格式化后的字段名称
func formatFieldName(field protoreflect.FieldDescriptor) string {
	return field.JSONName()
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

// schemaReferenceForMessage 获取消息的 Schema 引用
// 参数:
//   - message: 消息描述符
//
// 返回值:
//   - 消息的 Schema 引用
func schemaReferenceForMessage(message protoreflect.MessageDescriptor) string {
	schemaName := formatMessageName(message)
	return "#/components/schemas/" + schemaName
}

// filterCommentString 过滤注释字符串
// 参数:
//   - comment: 注释文本
//
// 返回值:
//   - 过滤后的注释文本
func filterCommentString(comment protogen.Comments) string {
	return string(comment)
}

// splitLines 将字符串按行分割成切片
// 参数:
//   - s: 要分割的字符串
//
// 返回值:
//   - 分割后的行切片
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// joinLines 将字符串切片合并成一个字符串
// 参数:
//   - lines: 行切片
//   - separator: 分隔符
//
// 返回值:
//   - 合并后的字符串
func joinLines(lines []string, separator string) string {
	return strings.Join(lines, separator)
}

// trimSpace 去除字符串前后的空白字符
// 参数:
//   - s: 要处理的字符串
//
// 返回值:
//   - 处理后的字符串
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// hasPrefix 检查字符串是否以指定前缀开头
// 参数:
//   - s: 要检查的字符串
//   - prefix: 前缀
//
// 返回值:
//   - 如果s以prefix开头则返回true，否则返回false
func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// hasSuffix 检查字符串是否以指定后缀结尾
// 参数:
//   - s: 要检查的字符串
//   - suffix: 后缀
//
// 返回值:
//   - 如果s以suffix结尾则返回true，否则返回false
func hasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// trimPrefix 去除字符串的前缀
// 参数:
//   - s: 要处理的字符串
//   - prefix: 要去除的前缀
//
// 返回值:
//   - 处理后的字符串
func trimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// trimSuffix 去除字符串的后缀
// 参数:
//   - s: 要处理的字符串
//   - suffix: 要去除的后缀
//
// 返回值:
//   - 处理后的字符串
func trimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// findInSlice 在切片中查找元素
// 参数:
//   - slice: 要搜索的切片
//   - predicate: 用于确定元素是否匹配的函数
//
// 返回值:
//   - 第一个匹配元素的索引，如果没有找到则返回-1
//   - 找到的元素
//   - 是否找到
func findInSlice[T any](slice []T, predicate func(T) bool) (int, T, bool) {
	for i, item := range slice {
		if predicate(item) {
			return i, item, true
		}
	}
	var zero T
	return -1, zero, false
}
