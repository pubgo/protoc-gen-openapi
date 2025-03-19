package pure

import (
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/protoc-gen-openapi/generator/model"
	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
)

// HTTPRule 表示一个 HTTP 规则
type HTTPRule struct {
	Path      string
	Method    string
	Body      string
	IsCustom  bool
	IsUnknown bool
}

// BuildHTTPRule 从 annotations.HttpRule 构建 HTTPRule
func BuildHTTPRule(rule *annotations.HttpRule) HTTPRule {
	result := HTTPRule{
		Body: rule.Body,
	}

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		result.Path = pattern.Get
		result.Method = "GET"
	case *annotations.HttpRule_Post:
		result.Path = pattern.Post
		result.Method = "POST"
	case *annotations.HttpRule_Put:
		result.Path = pattern.Put
		result.Method = "PUT"
	case *annotations.HttpRule_Delete:
		result.Path = pattern.Delete
		result.Method = "DELETE"
	case *annotations.HttpRule_Patch:
		result.Path = pattern.Patch
		result.Method = "PATCH"
	case *annotations.HttpRule_Custom:
		result.Path = "custom-unsupported"
		result.IsCustom = true
	default:
		result.Path = "unknown-unsupported"
		result.IsUnknown = true
	}

	return result
}

// BuildHTTPRules 从方法选项中提取所有 HTTP 规则
func BuildHTTPRules(method *protogen.Method) []*annotations.HttpRule {
	rules := make([]*annotations.HttpRule, 0)

	extHTTP := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
	if extHTTP != nil && extHTTP != annotations.E_Http.InterfaceOf(annotations.E_Http.Zero()) {
		rule := extHTTP.(*annotations.HttpRule)
		rules = append(rules, rule)
		rules = append(rules, rule.AdditionalBindings...)
	}

	return rules
}

// MergeServiceExtensions 合并服务级别的扩展到操作中
func MergeServiceExtensions(op *v3.Operation, extService *model.Service) {
	if extService == nil {
		return
	}

	op.Parameters = append(op.Parameters, extService.Parameters...)
	op.SpecificationExtension = append(op.SpecificationExtension, extService.SpecificationExtension...)
	op.Tags = append(op.Tags, extService.Tags...)
	op.Servers = append(op.Servers, extService.Servers...)
	op.Security = append(op.Security, extService.Security...)
	if extService.ExternalDocs != nil {
		proto.Merge(op.ExternalDocs, extService.ExternalDocs)
	}
}

// ProcessParameters 处理参数，确保 header 参数有正确的 schema
func ProcessParameters(parameters []*v3.ParameterOrReference) []*v3.ParameterOrReference {
	for _, v := range parameters {
		if v.Oneof == nil {
			continue
		}

		switch v1 := v.Oneof.(type) {
		case *v3.ParameterOrReference_Parameter:
			p := v1.Parameter
			if p.In == "header" && p.Schema == nil {
				p.Schema = wellknown.NewStringSchema()
			}
		}
	}
	return parameters
}

// ProcessTags 处理标签，分离出规范扩展
func ProcessTags(op *v3.Operation) {
	var tags []string
	var extensions []*v3.NamedAny

	for _, v := range op.Tags {
		if strings.Contains(v, "=") {
			tagNames := strings.SplitN(v, "=", 2)
			extensions = append(extensions, &v3.NamedAny{
				Name: strings.TrimSpace(tagNames[0]),
				Value: &v3.Any{
					Yaml: strings.TrimSpace(tagNames[1]),
				},
			})
			continue
		}
		tags = append(tags, v)
	}

	// 去重扩展
	extMap := make(map[string]*v3.NamedAny)
	for _, v := range append(op.SpecificationExtension, extensions...) {
		extMap[v.Name] = v
	}

	op.SpecificationExtension = op.SpecificationExtension[:0]
	for _, v := range extMap {
		op.SpecificationExtension = append(op.SpecificationExtension, v)
	}

	op.Tags = tags
}

// BuildPath 构建单个路径
func BuildPath(
	doc *v3.Document,
	service *protogen.Service,
	method *protogen.Method,
	rule *annotations.HttpRule,
	buildOperationFunc func(
		*v3.Document,
		string,
		string,
		string,
		string,
		string,
		string,
		*protogen.Message,
		*protogen.Message,
	) (*v3.Operation, string),
) (*v3.Operation, string, string) {
	httpRule := BuildHTTPRule(rule)
	if httpRule.IsCustom || httpRule.IsUnknown || httpRule.Method == "" {
		return nil, "", ""
	}

	comment := method.Comments.Leading.String()
	operationID := service.GoName + "_" + method.GoName
	defaultHost := proto.GetExtension(service.Desc.Options(), annotations.E_DefaultHost).(string)

	op, path := buildOperationFunc(
		doc,
		operationID,
		service.GoName,
		comment,
		defaultHost,
		httpRule.Path,
		httpRule.Body,
		method.Input,
		method.Output,
	)

	return op, path, httpRule.Method
}

// BuildPaths 构建所有路径
func BuildPaths(
	doc *v3.Document,
	services []*protogen.Service,
	buildOperationFunc func(
		*v3.Document,
		string,
		string,
		string,
		string,
		string,
		string,
		*protogen.Message,
		*protogen.Message,
	) (*v3.Operation, string),
	addOperationFunc func(*v3.Document, *v3.Operation, string, string),
) []string {
	var processedServices []string

	for _, service := range services {
		extService, _ := proto.GetExtension(service.Desc.Options(), model.E_Service).(*model.Service)
		annotationsCount := 0

		for _, method := range service.Methods {
			rules := BuildHTTPRules(method)

			for _, rule := range rules {
				op, path, methodName := BuildPath(doc, service, method, rule, buildOperationFunc)
				if op == nil {
					continue
				}

				annotationsCount++

				// 合并服务级别的扩展
				MergeServiceExtensions(op, extService)

				// 合并方法级别的扩展
				if extOperation := proto.GetExtension(method.Desc.Options(), v3.E_Operation); extOperation != nil {
					proto.Merge(op, extOperation.(*v3.Operation))
				}

				// 处理参数
				op.Parameters = ProcessParameters(op.Parameters)

				// 处理标签和规范扩展
				ProcessTags(op)

				// 添加操作到文档
				addOperationFunc(doc, op, path, methodName)
			}
		}

		if annotationsCount > 0 {
			comment := service.Comments.Leading.String()
			doc.Tags = append(doc.Tags, &v3.Tag{Name: service.GoName, Description: comment})
			processedServices = append(processedServices, service.GoName)
		}
	}

	return processedServices
}
