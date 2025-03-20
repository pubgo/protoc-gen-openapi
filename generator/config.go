package generator

import (
	"fmt"
	"strings"
)

// Configuration 定义了生成器的配置选项
// 所有字段都是指针类型，允许区分未设置和显式设置为零值的情况
type Configuration struct {
	// Version 指定生成的 OpenAPI 文档的版本号
	Version *string

	// Title 指定生成的 OpenAPI 文档的标题
	Title *string

	// Description 指定生成的 OpenAPI 文档的描述信息
	Description *string

	// Naming 指定命名约定，可选值: "proto"（使用 protobuf 字段名）, "json"（使用 JSON 字段名）
	Naming *string

	// FQSchemaNaming 指定是否使用完全限定的模式名称
	// 若为 true，则会在模式名称前添加 proto 包名
	FQSchemaNaming *bool

	// EnumType 指定枚举类型的序列化方式
	// 可选值: "string"（使用字符串表示）, "integer"（使用整数表示）
	EnumType *string

	// CircularDepth 指定处理循环引用消息的最大深度
	CircularDepth *int

	// DefaultResponse 指定是否添加默认响应
	// 若为 true，则会自动为使用 google.rpc.Status 消息的操作添加默认响应
	// 对于使用 envoy 或 grpc-gateway 进行转码的用户很有用，因为它们使用此类型作为默认错误响应
	DefaultResponse *bool

	// OutputMode 指定输出生成模式
	// 可选值: "merged"（在输出文件夹中生成单个 openapi.yaml）
	// "source_relative"（在每个输入文件旁边生成单独的 [inputfile].openapi.yaml）
	OutputMode *string
}

// DefaultConfig 返回默认配置
// 返回值:
//   - 默认的配置实例
func DefaultConfig() Configuration {
	// 设置默认值
	version := "1.0.0"
	title := "API"
	description := "Generated API"
	naming := "proto"
	fqSchemaNaming := false
	enumType := "string"
	circularDepth := 2
	defaultResponse := true
	outputMode := "merged"

	return Configuration{
		Version:         &version,
		Title:           &title,
		Description:     &description,
		Naming:          &naming,
		FQSchemaNaming:  &fqSchemaNaming,
		EnumType:        &enumType,
		CircularDepth:   &circularDepth,
		DefaultResponse: &defaultResponse,
		OutputMode:      &outputMode,
	}
}

// Validate 验证配置是否有效
// 返回值:
//   - 验证错误，如果配置有效则返回 nil
func (c *Configuration) Validate() error {
	// 检查必填字段
	if err := c.validateRequiredFields(); err != nil {
		return err
	}

	// 检查枚举字段值
	if err := c.validateEnumFields(); err != nil {
		return err
	}

	// 检查数值范围
	if err := c.validateNumericRanges(); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields 验证必填字段
// 返回值:
//   - 验证错误，如果所有必填字段有效则返回 nil
func (c *Configuration) validateRequiredFields() error {
	if c.Version == nil || *c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Title == nil || *c.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

// validateEnumFields 验证枚举字段值
// 返回值:
//   - 验证错误，如果所有枚举字段有效则返回 nil
func (c *Configuration) validateEnumFields() error {
	// 验证 Naming 字段
	if c.Naming != nil && *c.Naming != "proto" && *c.Naming != "json" {
		return fmt.Errorf("naming must be either 'proto' or 'json'")
	}

	// 验证 EnumType 字段
	if c.EnumType != nil && *c.EnumType != "string" && *c.EnumType != "integer" {
		return fmt.Errorf("enumType must be either 'string' or 'integer'")
	}

	// 验证 OutputMode 字段
	if c.OutputMode != nil && *c.OutputMode != "merged" && *c.OutputMode != "source_relative" {
		return fmt.Errorf("outputMode must be either 'merged' or 'source_relative'")
	}

	return nil
}

// validateNumericRanges 验证数值范围
// 返回值:
//   - 验证错误，如果所有数值字段在有效范围内则返回 nil
func (c *Configuration) validateNumericRanges() error {
	if c.CircularDepth != nil && *c.CircularDepth < 1 {
		return fmt.Errorf("circularDepth must be greater than 0")
	}
	return nil
}

// Merge 合并两个配置，返回新的配置
// 参数:
//   - other: 要合并的其他配置
//
// 返回值:
//   - 合并后的配置
func (c *Configuration) Merge(other Configuration) Configuration {
	result := *c

	// 只覆盖非 nil 值
	if other.Version != nil {
		result.Version = other.Version
	}
	if other.Title != nil {
		result.Title = other.Title
	}
	if other.Description != nil {
		result.Description = other.Description
	}
	if other.Naming != nil {
		result.Naming = other.Naming
	}
	if other.FQSchemaNaming != nil {
		result.FQSchemaNaming = other.FQSchemaNaming
	}
	if other.EnumType != nil {
		result.EnumType = other.EnumType
	}
	if other.CircularDepth != nil {
		result.CircularDepth = other.CircularDepth
	}
	if other.DefaultResponse != nil {
		result.DefaultResponse = other.DefaultResponse
	}
	if other.OutputMode != nil {
		result.OutputMode = other.OutputMode
	}

	return result
}

// String 返回配置的字符串表示
// 返回值:
//   - 配置的字符串表示
func (c *Configuration) String() string {
	var parts []string

	// 添加非 nil 字段
	if c.Version != nil {
		parts = append(parts, fmt.Sprintf("Version: %s", *c.Version))
	}
	if c.Title != nil {
		parts = append(parts, fmt.Sprintf("Title: %s", *c.Title))
	}
	if c.Description != nil {
		parts = append(parts, fmt.Sprintf("Description: %s", *c.Description))
	}
	if c.Naming != nil {
		parts = append(parts, fmt.Sprintf("Naming: %s", *c.Naming))
	}
	if c.FQSchemaNaming != nil {
		parts = append(parts, fmt.Sprintf("FQSchemaNaming: %v", *c.FQSchemaNaming))
	}
	if c.EnumType != nil {
		parts = append(parts, fmt.Sprintf("EnumType: %s", *c.EnumType))
	}
	if c.CircularDepth != nil {
		parts = append(parts, fmt.Sprintf("CircularDepth: %d", *c.CircularDepth))
	}
	if c.DefaultResponse != nil {
		parts = append(parts, fmt.Sprintf("DefaultResponse: %v", *c.DefaultResponse))
	}
	if c.OutputMode != nil {
		parts = append(parts, fmt.Sprintf("OutputMode: %s", *c.OutputMode))
	}

	return "Configuration{" + strings.Join(parts, ", ") + "}"
}
