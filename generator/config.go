package generator

import (
	"fmt"
)

// Configuration 定义了生成器的配置选项
type Configuration struct {
	Version         *string
	Title           *string
	Description     *string
	Naming          *string
	FQSchemaNaming  *bool
	EnumType        *string
	CircularDepth   *int
	DefaultResponse *bool
	OutputMode      *string
}

// DefaultConfig 返回默认配置
func DefaultConfig() Configuration {
	version := "1.0.0"
	title := "API"
	description := "Generated API"
	naming := "proto"
	fqSchemaNaming := false
	enumType := "string"
	circularDepth := 2
	defaultResponse := true
	outputMode := "yaml"

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
func (c *Configuration) Validate() error {
	if c.Version == nil || *c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.Title == nil || *c.Title == "" {
		return fmt.Errorf("title is required")
	}
	if c.Naming != nil && *c.Naming != "proto" && *c.Naming != "json" {
		return fmt.Errorf("naming must be either 'proto' or 'json'")
	}
	if c.EnumType != nil && *c.EnumType != "string" && *c.EnumType != "integer" {
		return fmt.Errorf("enumType must be either 'string' or 'integer'")
	}
	if c.CircularDepth != nil && *c.CircularDepth < 1 {
		return fmt.Errorf("circularDepth must be greater than 0")
	}
	if c.OutputMode != nil && *c.OutputMode != "yaml" && *c.OutputMode != "json" {
		return fmt.Errorf("outputMode must be either 'yaml' or 'json'")
	}
	return nil
}

// Merge 合并两个配置，返回新的配置
func (c *Configuration) Merge(other Configuration) Configuration {
	result := *c

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
