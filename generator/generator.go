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
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/pubgo/protoc-gen-openapi/pkg/servicev3"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/genproto/googleapis/api/annotations"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
)

const (
	infoURL = "https://github.com/pubgo/protoc-gen-openapi"
)

// In order to dynamically add google.rpc.Status responses we need
// to know the message descriptors for google.rpc.Status as well
// as google.protobuf.Any.
var (
	statusProtoDesc   = (&statuspb.Status{}).ProtoReflect().Descriptor()
	anyProtoDesc      = (&anypb.Any{}).ProtoReflect().Descriptor()
	linterRulePattern = regexp.MustCompile(`\(-- .* --\)`)
	namedPathPattern  = regexp.MustCompile("{(.+)=(.+)}")
	pathPattern       = regexp.MustCompile("{([^=}]+)}")
)

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a transcoded Protocol Buffer service.
type OpenAPIv3Generator struct {
	conf   Configuration
	plugin *protogen.Plugin

	inputFiles       []*protogen.File
	reflect          *OpenAPIv3Reflector
	generatedSchemas []string // Names of schemas that have already been generated.
}

// NewOpenAPIv3Generator creates a new generator for a protoc plugin invocation.
func NewOpenAPIv3Generator(plugin *protogen.Plugin, conf Configuration, inputFiles []*protogen.File) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		conf:   conf,
		plugin: plugin,

		inputFiles:       inputFiles,
		reflect:          NewOpenAPIv3Reflector(conf),
		generatedSchemas: make([]string, 0),
	}
}

// Run runs the generator.
func (g *OpenAPIv3Generator) Run(outputFile *protogen.GeneratedFile) error {
	d := g.buildDocumentV3()
	bytes, err := d.YAMLValue("Generated with protoc-gen-openapi\n" + infoURL)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %s", err.Error())
	}
	if _, err = outputFile.Write(bytes); err != nil {
		return fmt.Errorf("failed to write yaml: %s", err.Error())
	}
	return nil
}

// buildDocumentV3 builds an OpenAPIv3 document for a plugin request.
func (g *OpenAPIv3Generator) buildDocumentV3() *v3.Document {
	// 初始化文档
	d := initializeDocument(g.conf)

	// 处理文件
	g.processFiles(d, g.inputFiles)

	// 处理服务信息
	g.processServiceInfo(d)

	// 处理服务器信息
	g.processServers(d)

	// 排序文档
	g.sortDocument(d)

	return d
}

// processFiles processes all files and adds their paths to the document
func (g *OpenAPIv3Generator) processFiles(d *v3.Document, inputFiles []*protogen.File) {
	for _, file := range inputFiles {
		if !file.Generate {
			continue
		}

		// Merge any `Document` annotations with the current
		extDocument := proto.GetExtension(file.Desc.Options(), v3.E_Document)
		if extDocument != nil {
			proto.Merge(d, extDocument.(*v3.Document))
		}

		g.addPathsToDocumentV3(d, file.Services)
	}
}

// processServiceInfo processes service information and adds it to the document
func (g *OpenAPIv3Generator) processServiceInfo(d *v3.Document) {
	// While we have required schemas left to generate, go through the files again
	// looking for the related message and adding them to the document if required.
	for len(g.reflect.requiredSchemas) > 0 {
		count := len(g.reflect.requiredSchemas)
		for _, file := range g.plugin.Files {
			g.addSchemasForMessagesToDocumentV3(d, file.Messages)
		}
		g.reflect.requiredSchemas = g.reflect.requiredSchemas[count:len(g.reflect.requiredSchemas)]
	}

	// If there is only 1 service, then use it's title for the
	// document, if the document is missing it.
	g.updateDocumentInfoFromTags(d)
}

// updateDocumentInfoFromTags updates document information from tags if needed
func (g *OpenAPIv3Generator) updateDocumentInfoFromTags(d *v3.Document) {
	if len(d.Tags) == 1 {
		if d.Info.Title == "" && d.Tags[0].Name != "" {
			d.Info.Title = d.Tags[0].Name + " API"
		}
		if d.Info.Description == "" {
			d.Info.Description = d.Tags[0].Description
		}
		d.Tags[0].Description = ""
	}
}

// processServers processes server information and adds it to the document
func (g *OpenAPIv3Generator) processServers(d *v3.Document) {
	var allServers []string

	// If paths methods has servers, but they're all the same, then move servers to path level
	for _, path := range d.Paths.Path {
		var servers []string

		// 处理不同HTTP方法的服务器信息
		servers = g.collectServerURLs(path, servers)
		allServers = g.appendUniqueServers(allServers, servers)

		// 如果路径下所有方法都具有相同的服务器，则将服务器信息移至路径级别
		g.moveServersToPathLevel(path, servers)
	}

	// 如果所有路径都使用相同的服务器，则将服务器信息移至文档级别
	g.moveServersToDocumentLevel(d, allServers)
}

// collectServerURLs collects server URLs from different HTTP methods in a path
func (g *OpenAPIv3Generator) collectServerURLs(path *v3.NamedPathItem, servers []string) []string {
	// Only 1 server will ever be set, per method, by the generator
	if path.Value.Get != nil && len(path.Value.Get.Servers) == 1 {
		servers = appendUnique(servers, path.Value.Get.Servers[0].Url)
	}

	if path.Value.Post != nil && len(path.Value.Post.Servers) == 1 {
		servers = appendUnique(servers, path.Value.Post.Servers[0].Url)
	}

	if path.Value.Put != nil && len(path.Value.Put.Servers) == 1 {
		servers = appendUnique(servers, path.Value.Put.Servers[0].Url)
	}

	if path.Value.Delete != nil && len(path.Value.Delete.Servers) == 1 {
		servers = appendUnique(servers, path.Value.Delete.Servers[0].Url)
	}

	if path.Value.Patch != nil && len(path.Value.Patch.Servers) == 1 {
		servers = appendUnique(servers, path.Value.Patch.Servers[0].Url)
	}

	return servers
}

// appendUniqueServers appends unique server URLs to allServers list
func (g *OpenAPIv3Generator) appendUniqueServers(allServers []string, servers []string) []string {
	for _, server := range servers {
		allServers = appendUnique(allServers, server)
	}
	return allServers
}

// moveServersToPathLevel moves servers to path level if all methods use the same servers
func (g *OpenAPIv3Generator) moveServersToPathLevel(path *v3.NamedPathItem, servers []string) {
	if len(servers) == 1 {
		allEqual := true

		if path.Value.Get != nil && (len(path.Value.Get.Servers) != 1 || path.Value.Get.Servers[0].Url != servers[0]) {
			allEqual = false
		}

		if path.Value.Post != nil && (len(path.Value.Post.Servers) != 1 || path.Value.Post.Servers[0].Url != servers[0]) {
			allEqual = false
		}

		if path.Value.Put != nil && (len(path.Value.Put.Servers) != 1 || path.Value.Put.Servers[0].Url != servers[0]) {
			allEqual = false
		}

		if path.Value.Delete != nil && (len(path.Value.Delete.Servers) != 1 || path.Value.Delete.Servers[0].Url != servers[0]) {
			allEqual = false
		}

		if path.Value.Patch != nil && (len(path.Value.Patch.Servers) != 1 || path.Value.Patch.Servers[0].Url != servers[0]) {
			allEqual = false
		}

		if allEqual {
			path.Value.Servers = []*v3.Server{{Url: servers[0]}}

			// 清除各个方法的服务器信息
			g.clearMethodServers(path)
		}
	}
}

// clearMethodServers clears server information from each HTTP method in a path
func (g *OpenAPIv3Generator) clearMethodServers(path *v3.NamedPathItem) {
	if path.Value.Get != nil {
		path.Value.Get.Servers = nil
	}
	if path.Value.Post != nil {
		path.Value.Post.Servers = nil
	}
	if path.Value.Put != nil {
		path.Value.Put.Servers = nil
	}
	if path.Value.Delete != nil {
		path.Value.Delete.Servers = nil
	}
	if path.Value.Patch != nil {
		path.Value.Patch.Servers = nil
	}
}

// moveServersToDocumentLevel moves servers to document level if all paths use the same servers
func (g *OpenAPIv3Generator) moveServersToDocumentLevel(d *v3.Document, allServers []string) {
	if len(allServers) == 1 {
		for _, path := range d.Paths.Path {
			if len(path.Value.Servers) == 1 && path.Value.Servers[0].Url == allServers[0] {
				path.Value.Servers = nil
			} else {
				return
			}
		}
		d.Servers = []*v3.Server{{Url: allServers[0]}}
	}
}

// sortDocument sorts the document tags and paths
func (g *OpenAPIv3Generator) sortDocument(d *v3.Document) {
	// Sort the tags.
	{
		pairs := d.Tags
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Tags = pairs
	}

	// Sort the paths.
	{
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}

	// Sort the schemas.
	{
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
	}
}

// filterCommentString removes linter rules from comments.
func filterCommentString(c protogen.Comments) string {
	comment := linterRulePattern.ReplaceAllString(string(c), "")
	return strings.TrimSpace(comment)
}

func findField(name string, inMessage *protogen.Message) *protogen.Field {
	for _, field := range inMessage.Fields {
		if string(field.Desc.Name()) == name || string(field.Desc.JSONName()) == name {
			return field
		}
	}

	return nil
}

func (g *OpenAPIv3Generator) findAndFormatFieldName(name string, inMessage *protogen.Message) string {
	field := findField(name, inMessage)
	if field != nil {
		return formatFieldName(*g.conf.Naming, field.Desc)
	}

	return name
}

// Note that fields which are mapped to URL query parameters must have a primitive type
// or a repeated primitive type or a non-repeated message type.
// In the case of a repeated type, the parameter can be repeated in the URL as ...?param=A&param=B.
// In the case of a message type, each field of the message is mapped to a separate parameter,
// such as ...?foo.a=A&foo.b=B&foo.c=C.
// There are exceptions:
// - for wrapper types it will use the same representation as the wrapped primitive type in JSON
// - for google.protobuf.timestamp type it will be serialized as a string
//
// maps, Struct and Empty can NOT be used
// messages can have any number of sub messages - including circular (e.g. sub.subsub.sub.subsub.id)

// buildQueryParamsV3 builds query parameters for a field
func (g *OpenAPIv3Generator) buildQueryParamsV3(field *protogen.Field) []*v3.ParameterOrReference {
	// 创建一个空的深度映射，用于跟踪递归深度
	depths := make(map[string]int)
	return g._buildQueryParamsV3(field, depths)
}

// _buildQueryParamsV3 is the internal implementation of buildQueryParamsV3
func (g *OpenAPIv3Generator) _buildQueryParamsV3(field *protogen.Field, depths map[string]int) []*v3.ParameterOrReference {
	fieldName := string(field.Desc.Name())
	jsonName := field.Desc.JSONName()

	// 获取字段描述和默认值
	fieldDescription, defaultValue := g.getFieldDescriptionAndDefault(field)

	// 处理不同类型的字段
	if field.Desc.Kind() == protoreflect.MessageKind && !field.Desc.IsMap() {
		return g.processMessageField(field, fieldName, jsonName, fieldDescription, defaultValue, depths)
	} else {
		// 对于基本类型和Map类型的字段，直接创建参数
		return []*v3.ParameterOrReference{
			g.createQueryParameter(jsonName, fieldDescription, defaultValue, g.reflect.schemaOrReferenceForField(field, nil)),
		}
	}
}

// getFieldDescriptionAndDefault extracts field description and default value
func (g *OpenAPIv3Generator) getFieldDescriptionAndDefault(field *protogen.Field) (string, *v3.Any) {
	var fieldDescription string
	var defaultValue *v3.Any

	comment := filterCommentString(field.Comments.Leading)
	commentLines := strings.Split(comment, "\n")
	fieldDescription = commentLines[0]

	// 检查是否有默认值标记
	for _, line := range commentLines {
		if strings.Contains(line, "Default:") {
			parts := strings.SplitN(line, "Default:", 2)
			if len(parts) > 1 {
				defaultValue = &v3.Any{
					Yaml: strings.TrimSpace(parts[1]),
				}
				break
			}
		}
	}

	return fieldDescription, defaultValue
}

// processMessageField processes a message field for query parameters
func (g *OpenAPIv3Generator) processMessageField(
	field *protogen.Field,
	fieldName string,
	jsonName string,
	fieldDescription string,
	defaultValue *v3.Any,
	depths map[string]int,
) []*v3.ParameterOrReference {
	// 检查递归深度
	typeName := string(field.Message.Desc.FullName())
	currentDepth, ok := depths[typeName]
	if !ok {
		currentDepth = 0
	}

	// 如果达到最大递归深度，则停止
	if currentDepth >= *g.conf.CircularDepth {
		return []*v3.ParameterOrReference{}
	}

	// 递增深度计数
	depths[typeName] = currentDepth + 1

	// 对于特殊的类型，给予特殊处理
	if typeName == "google.protobuf.Timestamp" ||
		typeName == "google.type.Date" ||
		typeName == "google.type.DateTime" {
		return []*v3.ParameterOrReference{
			g.createQueryParameter(jsonName, fieldDescription, defaultValue, g.reflect.schemaOrReferenceForField(field, nil)),
		}
	}

	// 处理嵌套消息类型
	var result []*v3.ParameterOrReference

	// 遍历嵌套消息的所有字段
	for _, subField := range field.Message.Fields {
		subJsonName := subField.Desc.JSONName()
		paramName := jsonName + "." + subJsonName

		// 递归处理每个字段
		if subField.Desc.Kind() == protoreflect.MessageKind && !subField.Desc.IsMap() {
			// 对于嵌套消息，递归调用处理函数
			for _, p := range g._buildQueryParamsV3(subField, depths) {
				if param := g.getParameterFromReference(p); param != nil {
					param.Name = jsonName + "." + param.Name
					result = append(result, p)
				}
			}
		} else {
			// 对于基本类型，创建参数
			subFieldDescription, _ := g.getFieldDescriptionAndDefault(subField)
			description := fieldDescription
			if subFieldDescription != "" {
				description = subFieldDescription
			}

			result = append(result, g.createQueryParameter(
				paramName,
				description,
				nil, // 不设置默认值
				g.reflect.schemaOrReferenceForField(subField, nil),
			))
		}
	}

	// 减少深度计数
	depths[typeName] = currentDepth

	return result
}

// getParameterFromReference extracts parameter from parameter reference
func (g *OpenAPIv3Generator) getParameterFromReference(p *v3.ParameterOrReference) *v3.Parameter {
	if p.Oneof == nil {
		return nil
	}

	if param, ok := p.Oneof.(*v3.ParameterOrReference_Parameter); ok {
		return param.Parameter
	}

	return nil
}

// createQueryParameter creates a query parameter
func (g *OpenAPIv3Generator) createQueryParameter(
	name string,
	description string,
	defaultValue *v3.Any,
	schema *v3.SchemaOrReference,
) *v3.ParameterOrReference {
	parameter := &v3.Parameter{
		Name:        name,
		In:          "query",
		Description: description,
		Required:    false,
		Schema:      schema,
	}

	// 如果有默认值，添加到扩展属性中
	if defaultValue != nil {
		parameter.SpecificationExtension = append(parameter.SpecificationExtension, &v3.NamedAny{
			Name:  "x-default",
			Value: defaultValue,
		})
	}

	return &v3.ParameterOrReference{
		Oneof: &v3.ParameterOrReference_Parameter{
			Parameter: parameter,
		},
	}
}

// buildOperation builds an operation
func (g *OpenAPIv3Generator) buildOperation(
	d *v3.Document,
	operationID string,
	tagName string,
	description string,
	defaultHost string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
) (*v3.Operation, string) {
	// 构建路径参数
	parameters, coveredParameters, newPath := g.buildPathParameters(path, inputMessage)

	// 添加请求体参数到已覆盖列表
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}

	// 添加查询参数
	parameters = g.addQueryParameters(parameters, coveredParameters, bodyField, inputMessage)

	// 创建操作对象
	op := &v3.Operation{
		Tags:        []string{tagName},
		Description: description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   g.buildResponses(outputMessage, *g.conf.DefaultResponse, d),
		Servers:     BuildServer(defaultHost),
		RequestBody: g.buildRequestBody(bodyField, inputMessage),
	}

	return op, newPath
}

// addQueryParameters adds query parameters to the parameter list
func (g *OpenAPIv3Generator) addQueryParameters(
	parameters []*v3.ParameterOrReference,
	coveredParameters []string,
	bodyField string,
	inputMessage *protogen.Message,
) []*v3.ParameterOrReference {
	if bodyField != "*" && string(inputMessage.Desc.FullName()) != "google.api.HttpBody" {
		for _, field := range inputMessage.Fields {
			fieldName := string(field.Desc.Name())
			if !contains(coveredParameters, fieldName) && fieldName != bodyField {
				fieldParams := g.buildQueryParamsV3(field)
				parameters = append(parameters, fieldParams...)
			}
		}
	}
	return parameters
}

// buildPathParameters builds path parameters
func (g *OpenAPIv3Generator) buildPathParameters(path string, inputMessage *protogen.Message) ([]*v3.ParameterOrReference, []string, string) {
	var parameters []*v3.ParameterOrReference
	coveredParameters := make([]string, 0)

	// 处理简单路径参数 {id}
	parameters, coveredParameters, path = g.processSimplePathParameters(path, parameters, coveredParameters, inputMessage)

	// 处理命名路径参数 {name=shelves/*}
	parameters, coveredParameters, path = g.processNamedPathParameters(path, parameters, coveredParameters, inputMessage)

	return parameters, coveredParameters, path
}

// processSimplePathParameters processes simple path parameters like {id}
func (g *OpenAPIv3Generator) processSimplePathParameters(
	path string,
	parameters []*v3.ParameterOrReference,
	coveredParameters []string,
	inputMessage *protogen.Message,
) ([]*v3.ParameterOrReference, []string, string) {
	if allMatches := pathPattern.FindAllStringSubmatch(path, -1); allMatches != nil {
		for _, matches := range allMatches {
			coveredParameters = append(coveredParameters, matches[1])
			pathParameter := g.findAndFormatFieldName(matches[1], inputMessage)
			path = strings.Replace(path, matches[1], pathParameter, 1)

			var fieldSchema *v3.SchemaOrReference
			var fieldDescription string
			field := findField(pathParameter, inputMessage)
			if field != nil {
				fieldSchema = g.reflect.schemaOrReferenceForField(field, nil)
				fieldDescription = filterCommentString(field.Comments.Leading)
			} else {
				fieldSchema = &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Type: "string",
						},
					},
				}
			}

			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        pathParameter,
							In:          "path",
							Description: fieldDescription,
							Required:    true,
							Schema:      fieldSchema,
						},
					},
				})
		}
	}
	return parameters, coveredParameters, path
}

// processNamedPathParameters processes named path parameters like {name=shelves/*}
func (g *OpenAPIv3Generator) processNamedPathParameters(
	path string,
	parameters []*v3.ParameterOrReference,
	coveredParameters []string,
	inputMessage *protogen.Message,
) ([]*v3.ParameterOrReference, []string, string) {
	if matches := namedPathPattern.FindStringSubmatch(path); matches != nil {
		namedPathParameters := make([]string, 0)
		coveredParameters = append(coveredParameters, matches[1])

		starredPath := matches[2]
		parts := strings.Split(starredPath, "/")
		for i := 0; i < len(parts)-1; i += 2 {
			section := parts[i]
			namedPathParameter := g.findAndFormatFieldName(section, inputMessage)
			namedPathParameter = singular(namedPathParameter)
			parts[i+1] = "{" + namedPathParameter + "}"
			namedPathParameters = append(namedPathParameters, namedPathParameter)
		}

		newPath := strings.Join(parts, "/")
		path = strings.Replace(path, matches[0], newPath, 1)

		for _, namedPathParameter := range namedPathParameters {
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        namedPathParameter,
							In:          "path",
							Required:    true,
							Description: "The " + namedPathParameter + " id.",
							Schema: &v3.SchemaOrReference{
								Oneof: &v3.SchemaOrReference_Schema{
									Schema: &v3.Schema{
										Type: "string",
									},
								},
							},
						},
					},
				})
		}
	}
	return parameters, coveredParameters, path
}

// buildOperationV3 constructs an operation for a set of values.
func (g *OpenAPIv3Generator) buildOperationV3(
	d *v3.Document,
	operationID string,
	tagName string,
	description string,
	defaultHost string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
) (*v3.Operation, string) {
	return g.buildOperation(d, operationID, tagName, description, defaultHost, path, bodyField, inputMessage, outputMessage)
}

// addOperationToDocumentV3 adds an operation to the specified path/method.
func addOperationToDocumentV3(d *v3.Document, op *v3.Operation, path, methodName string) {
	var selectedPathItem *v3.NamedPathItem
	for _, namedPathItem := range d.Paths.Path {
		if namedPathItem.Name == path {
			selectedPathItem = namedPathItem
			break
		}
	}
	// If we get here, we need to create a path item.
	if selectedPathItem == nil {
		selectedPathItem = &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
		d.Paths.Path = append(d.Paths.Path, selectedPathItem)
	}
	// Set the operation on the specified method.
	switch methodName {
	case "GET":
		selectedPathItem.Value.Get = op
	case "POST":
		selectedPathItem.Value.Post = op
	case "PUT":
		selectedPathItem.Value.Put = op
	case "DELETE":
		selectedPathItem.Value.Delete = op
	case "PATCH":
		selectedPathItem.Value.Patch = op
	}
}

// addPathsToDocumentV3 adds paths to document from services
func (g *OpenAPIv3Generator) addPathsToDocumentV3(d *v3.Document, services []*protogen.Service) {
	for _, service := range services {
		serviceExtension := extractServiceExtension(service)
		annotationsCount := 0

		for _, method := range service.Methods {
			annotationsCount += g.processMethodAnnotations(d, service, method, serviceExtension)
		}

		if annotationsCount > 0 {
			addServiceTag(d, service)
		}
	}
}

// extractServiceExtension extracts service extensions from service options
func extractServiceExtension(service *protogen.Service) *servicev3.Service {
	serviceExtension, _ := proto.GetExtension(service.Desc.Options(), servicev3.E_Service).(*servicev3.Service)
	return serviceExtension
}

// processMethodAnnotations processes all HTTP annotations for a method
func (g *OpenAPIv3Generator) processMethodAnnotations(d *v3.Document, service *protogen.Service, method *protogen.Method, serviceExtension *servicev3.Service) int {
	count := 0
	rules := extractHttpRules(method)

	for _, rule := range rules {
		count += g.processHttpRule(d, service, method, rule, serviceExtension)
	}

	return count
}

// extractHttpRules extracts HTTP rules from method options
func extractHttpRules(method *protogen.Method) []*annotations.HttpRule {
	rules := make([]*annotations.HttpRule, 0)

	extHTTP := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
	if extHTTP != nil && extHTTP != annotations.E_Http.InterfaceOf(annotations.E_Http.Zero()) {
		rule := extHTTP.(*annotations.HttpRule)
		rules = append(rules, rule)
		rules = append(rules, rule.AdditionalBindings...)
	}

	return rules
}

// processHttpRule processes a single HTTP rule for a method
func (g *OpenAPIv3Generator) processHttpRule(d *v3.Document, service *protogen.Service, method *protogen.Method, rule *annotations.HttpRule, serviceExtension *servicev3.Service) int {
	var methodName string
	var httpRule HTTPRule

	httpRule = buildHTTPRule(rule)
	if httpRule.IsCustom || httpRule.IsUnknown {
		return 0
	}

	methodName = httpRule.Method
	if methodName == "" {
		return 0
	}

	op, newPath := g.buildOperationForMethod(d, service, method, rule, httpRule)
	if op == nil {
		return 0
	}

	// 合并服务级别的扩展
	mergeServiceExtensions(op, serviceExtension)

	// 合并方法级别的扩展
	mergeMethodExtensions(op, method)

	// 处理参数
	processOperationParameters(op)

	// 处理标签和规范扩展
	processOperationTags(op)

	// 添加操作到文档
	addOperationToDocumentV3(d, op, newPath, methodName)

	return 1
}

// buildHTTPRule builds an HTTPRule from an annotations.HttpRule
func buildHTTPRule(rule *annotations.HttpRule) HTTPRule {
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

// buildOperationForMethod builds an operation for a method
func (g *OpenAPIv3Generator) buildOperationForMethod(d *v3.Document, service *protogen.Service, method *protogen.Method, rule *annotations.HttpRule, httpRule HTTPRule) (*v3.Operation, string) {
	comment := method.Comments.Leading.String()
	operationID := service.GoName + "_" + method.GoName
	defaultHost := proto.GetExtension(service.Desc.Options(), annotations.E_DefaultHost).(string)

	return g.buildOperation(
		d,
		operationID,
		service.GoName,
		comment,
		defaultHost,
		httpRule.Path,
		httpRule.Body,
		method.Input,
		method.Output,
	)
}

// mergeServiceExtensions merges service extensions into an operation
func mergeServiceExtensions(op *v3.Operation, serviceExtension *servicev3.Service) {
	if serviceExtension == nil {
		return
	}

	op.Parameters = append(op.Parameters, serviceExtension.Parameters...)
	op.SpecificationExtension = append(op.SpecificationExtension, serviceExtension.SpecificationExtension...)
	op.Tags = append(op.Tags, serviceExtension.Tags...)
	op.Servers = append(op.Servers, serviceExtension.Servers...)
	op.Security = append(op.Security, serviceExtension.Security...)
	if serviceExtension.ExternalDocs != nil {
		proto.Merge(op.ExternalDocs, serviceExtension.ExternalDocs)
	}
}

// mergeMethodExtensions merges method extensions into an operation
func mergeMethodExtensions(op *v3.Operation, method *protogen.Method) {
	if extOperation := proto.GetExtension(method.Desc.Options(), v3.E_Operation); extOperation != nil {
		proto.Merge(op, extOperation.(*v3.Operation))
	}
}

// processOperationParameters processes operation parameters
func processOperationParameters(op *v3.Operation) {
	for _, v := range op.Parameters {
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
}

// processOperationTags processes operation tags and specification extensions
func processOperationTags(op *v3.Operation) {
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

	// Map extensions to ensure unique keys
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

// addServiceTag adds a service tag to the document
func addServiceTag(d *v3.Document, service *protogen.Service) {
	comment := service.Comments.Leading.String()
	d.Tags = append(d.Tags, &v3.Tag{Name: service.GoName, Description: comment})
}

// HTTPRule represents an HTTP rule
type HTTPRule struct {
	Path      string
	Method    string
	Body      string
	IsCustom  bool
	IsUnknown bool
}

// addSchemaForMessageToDocumentV3 adds the schema to the document if required
func (g *OpenAPIv3Generator) addSchemaToDocumentV3(d *v3.Document, schema *v3.NamedSchemaOrReference) {
	if contains(g.generatedSchemas, schema.Name) {
		return
	}
	g.generatedSchemas = append(g.generatedSchemas, schema.Name)
	d.Components.Schemas.AdditionalProperties = append(d.Components.Schemas.AdditionalProperties, schema)
}

// addSchemasForMessagesToDocumentV3 adds info from one file descriptor.
func (g *OpenAPIv3Generator) addSchemasForMessagesToDocumentV3(d *v3.Document, messages []*protogen.Message) {
	// For each message, generate a definition.
	for _, message := range messages {
		if message.Messages != nil {
			g.addSchemasForMessagesToDocumentV3(d, message.Messages)
		}

		schemaName := formatMessageName(&g.conf, message.Desc)

		// Only generate this if we need it and haven't already generated it.
		if !contains(g.reflect.requiredSchemas, schemaName) ||
			contains(g.generatedSchemas, schemaName) {
			continue
		}

		typeName := fullMessageTypeName(message.Desc)
		messageDescription := filterCommentString(message.Comments.Leading)

		// `google.protobuf.Value` and `google.protobuf.Any` have special JSON transcoding
		// so we can't just reflect on the message descriptor.
		if typeName == ".google.protobuf.Value" {
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleProtobufValueSchema(schemaName))
			continue
		} else if typeName == ".google.protobuf.Any" {
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleProtobufAnySchema(schemaName))
			continue
		} else if typeName == ".google.rpc.Status" {
			anySchemaName := formatMessageName(&g.conf, anyProtoDesc)
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleProtobufAnySchema(anySchemaName))
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleRpcStatusSchema(schemaName, anySchemaName))
			continue
		}

		// Build an array holding the fields of the message.
		definitionProperties := &v3.Properties{
			AdditionalProperties: make([]*v3.NamedSchemaOrReference, 0),
		}

		var required []string
		for _, field := range message.Fields {
			// Get the field description from the comments.
			description := filterCommentString(field.Comments.Leading)
			// Check the field annotations to see if this is a readonly or writeonly field.
			inputOnly := false
			outputOnly := false
			extension := proto.GetExtension(field.Desc.Options(), annotations.E_FieldBehavior)
			if extension != nil {
				switch v := extension.(type) {
				case []annotations.FieldBehavior:
					for _, vv := range v {
						switch vv {
						case annotations.FieldBehavior_OUTPUT_ONLY:
							outputOnly = true
						case annotations.FieldBehavior_INPUT_ONLY:
							inputOnly = true
						case annotations.FieldBehavior_REQUIRED:
							required = append(required, formatFieldName(*g.conf.Naming, field.Desc))
						}
					}
				default:
					log.Printf("unsupported extension type %T", extension)
				}
			}

			// The field is either described by a reference or a schema.
			fieldSchema := g.reflect.schemaOrReferenceForField(field, nil)
			if fieldSchema == nil {
				continue
			}

			// If this field has siblings and is a $ref now, create a new schema use `allOf` to wrap it
			wrapperNeeded := inputOnly || outputOnly || description != ""
			if wrapperNeeded {
				if _, ok := fieldSchema.Oneof.(*v3.SchemaOrReference_Reference); ok {
					fieldSchema = &v3.SchemaOrReference{Oneof: &v3.SchemaOrReference_Schema{Schema: &v3.Schema{
						AllOf: []*v3.SchemaOrReference{fieldSchema},
					}}}
				}
			}

			if schema, ok := fieldSchema.Oneof.(*v3.SchemaOrReference_Schema); ok {
				if schema.Schema.Description == "" {
					schema.Schema.Description = description
				}
				schema.Schema.ReadOnly = outputOnly
				schema.Schema.WriteOnly = inputOnly
				if opt, ok := field.Desc.Options().(*descriptorpb.FieldOptions); ok && opt != nil {
					if opt.Deprecated != nil && *opt.Deprecated {
						schema.Schema.Deprecated = *opt.Deprecated
					}
				}

				// Merge any `Property` annotations with the current
				extProperty := proto.GetExtension(field.Desc.Options(), v3.E_Property)
				if extProperty != nil {
					proto.Merge(schema.Schema, extProperty.(*v3.Schema))
				}
			}

			definitionProperties.AdditionalProperties = append(
				definitionProperties.AdditionalProperties,
				&v3.NamedSchemaOrReference{
					Name:  formatFieldName(*g.conf.Naming, field.Desc),
					Value: fieldSchema,
				},
			)
		}

		schema := &v3.Schema{
			Type:        "object",
			Description: messageDescription,
			Properties:  definitionProperties,
			Required:    required,
		}

		// Merge any `Schema` annotations with the current
		extSchema := proto.GetExtension(message.Desc.Options(), v3.E_Schema)
		if extSchema != nil {
			proto.Merge(schema, extSchema.(*v3.Schema))
		}

		// Add the schema to the components.schema list.
		g.addSchemaToDocumentV3(d, &v3.NamedSchemaOrReference{
			Name: schemaName,
			Value: &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: schema,
				},
			},
		})
	}
}
