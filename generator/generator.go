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
	statusProtoDesc = (&statuspb.Status{}).ProtoReflect().Descriptor()
	anyProtoDesc    = (&anypb.Any{}).ProtoReflect().Descriptor()
)

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a transcoded Protocol Buffer service.
type OpenAPIv3Generator struct {
	conf   Configuration
	plugin *protogen.Plugin

	inputFiles        []*protogen.File
	reflect           *OpenAPIv3Reflector
	generatedSchemas  []string // Names of schemas that have already been generated.
	linterRulePattern *regexp.Regexp
	pathPattern       *regexp.Regexp
	namedPathPattern  *regexp.Regexp
}

// NewOpenAPIv3Generator creates a new generator for a protoc plugin invocation.
func NewOpenAPIv3Generator(plugin *protogen.Plugin, conf Configuration, inputFiles []*protogen.File) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		conf:   conf,
		plugin: plugin,

		inputFiles:        inputFiles,
		reflect:           NewOpenAPIv3Reflector(conf),
		generatedSchemas:  make([]string, 0),
		linterRulePattern: regexp.MustCompile(`\(-- .* --\)`),
		pathPattern:       regexp.MustCompile("{([^=}]+)}"),
		namedPathPattern:  regexp.MustCompile("{(.+)=(.+)}"),
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
	d := g.initializeDocument()

	// 处理文件
	g.processFiles(d)

	// 处理服务信息
	g.processServiceInfo(d)

	// 处理服务器信息
	g.processServers(d)

	// 排序文档
	g.sortDocument(d)

	return d
}

// initializeDocument initializes the document with basic information
func (g *OpenAPIv3Generator) initializeDocument() *v3.Document {
	d := &v3.Document{}

	d.Openapi = "3.0.3"
	d.Info = &v3.Info{
		Version:     *g.conf.Version,
		Title:       *g.conf.Title,
		Description: *g.conf.Description,
	}

	d.Paths = &v3.Paths{}
	d.Components = &v3.Components{
		Schemas: &v3.SchemasOrReferences{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		},
	}

	return d
}

// processFiles processes all files and adds their paths to the document
func (g *OpenAPIv3Generator) processFiles(d *v3.Document) {
	for _, file := range g.inputFiles {
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
		// Only 1 server will ever be set, per method, by the generator

		if path.Value.Get != nil && len(path.Value.Get.Servers) == 1 {
			servers = appendUnique(servers, path.Value.Get.Servers[0].Url)
			allServers = appendUnique(servers, path.Value.Get.Servers[0].Url)
		}

		if path.Value.Post != nil && len(path.Value.Post.Servers) == 1 {
			servers = appendUnique(servers, path.Value.Post.Servers[0].Url)
			allServers = appendUnique(servers, path.Value.Post.Servers[0].Url)
		}

		if path.Value.Put != nil && len(path.Value.Put.Servers) == 1 {
			servers = appendUnique(servers, path.Value.Put.Servers[0].Url)
			allServers = appendUnique(servers, path.Value.Put.Servers[0].Url)
		}

		if path.Value.Delete != nil && len(path.Value.Delete.Servers) == 1 {
			servers = appendUnique(servers, path.Value.Delete.Servers[0].Url)
			allServers = appendUnique(servers, path.Value.Delete.Servers[0].Url)
		}

		if path.Value.Patch != nil && len(path.Value.Patch.Servers) == 1 {
			servers = appendUnique(servers, path.Value.Patch.Servers[0].Url)
			allServers = appendUnique(servers, path.Value.Patch.Servers[0].Url)
		}

		if len(servers) == 1 {
			path.Value.Servers = []*v3.Server{{Url: servers[0]}}

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
	}

	// Set all servers on API level
	if len(allServers) > 0 {
		d.Servers = []*v3.Server{}
		for _, server := range allServers {
			d.Servers = append(d.Servers, &v3.Server{Url: server})
		}
	}

	// If there is only 1 server, we can safely remove all path level servers
	if len(allServers) == 1 {
		for _, path := range d.Paths.Path {
			path.Value.Servers = nil
		}
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
func (g *OpenAPIv3Generator) filterCommentString(c protogen.Comments) string {
	comment := g.linterRulePattern.ReplaceAllString(string(c), "")
	return strings.TrimSpace(comment)
}

func (g *OpenAPIv3Generator) findField(name string, inMessage *protogen.Message) *protogen.Field {
	for _, field := range inMessage.Fields {
		if string(field.Desc.Name()) == name || string(field.Desc.JSONName()) == name {
			return field
		}
	}

	return nil
}

func (g *OpenAPIv3Generator) findAndFormatFieldName(name string, inMessage *protogen.Message) string {
	field := g.findField(name, inMessage)
	if field != nil {
		return g.reflect.formatFieldName(field.Desc)
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

// buildQueryParamsV3 extracts any valid query params, including sub and recursive messages
func (g *OpenAPIv3Generator) buildQueryParamsV3(field *protogen.Field) []*v3.ParameterOrReference {
	depths := map[string]int{}
	return g._buildQueryParamsV3(field, depths)
}

// depths are used to keep track of how many times a message's fields has been seen
func (g *OpenAPIv3Generator) _buildQueryParamsV3(field *protogen.Field, depths map[string]int) []*v3.ParameterOrReference {
	var parameters []*v3.ParameterOrReference
	queryFieldName := g.reflect.formatFieldName(field.Desc)
	fieldDescription := g.filterCommentString(field.Comments.Leading)

	if field.Desc.IsMap() {
		// Map types are not allowed in query parameteres
		return parameters
	}

	if field.Desc.Kind() == protoreflect.MessageKind {
		typeName := fullMessageTypeName(field.Desc.Message())
		switch typeName {
		case ".google.protobuf.Value":
			fieldSchema := g.reflect.schemaOrReferenceForField(field, nil)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters

		case ".google.protobuf.BoolValue", ".google.protobuf.BytesValue", ".google.protobuf.Int32Value", ".google.protobuf.UInt32Value",
			".google.protobuf.StringValue", ".google.protobuf.Int64Value", ".google.protobuf.UInt64Value", ".google.protobuf.FloatValue",
			".google.protobuf.DoubleValue":
			fieldSchema := g.reflect.schemaOrReferenceForField(field, nil)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters

		case ".google.protobuf.Timestamp":
			fieldSchema := g.reflect.schemaOrReferenceForMessage(field.Message.Desc)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters
		case ".google.protobuf.Duration":
			fieldSchema := g.reflect.schemaOrReferenceForMessage(field.Message.Desc)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters
		}

		if field.Desc.IsList() {
			// Only non-repeated message types are valid
			return parameters
		}

		// Represent field masks directly as strings (don't expand them).
		if typeName == ".google.protobuf.FieldMask" {
			fieldSchema := g.reflect.schemaOrReferenceForField(field, nil)
			parameters = append(parameters,
				&v3.ParameterOrReference{
					Oneof: &v3.ParameterOrReference_Parameter{
						Parameter: &v3.Parameter{
							Name:        queryFieldName,
							In:          "query",
							Description: fieldDescription,
							Required:    false,
							Schema:      fieldSchema,
						},
					},
				})
			return parameters
		}

		// Sub messages are allowed, even circular, as long as the final type is a primitive.
		// Go through each of the sub message fields
		for _, subField := range field.Message.Fields {
			subFieldFullName := string(subField.Desc.FullName())
			seen, ok := depths[subFieldFullName]
			if !ok {
				depths[subFieldFullName] = 0
			}

			if seen < *g.conf.CircularDepth {
				depths[subFieldFullName]++
				subParams := g._buildQueryParamsV3(subField, depths)
				for _, subParam := range subParams {
					if param, ok := subParam.Oneof.(*v3.ParameterOrReference_Parameter); ok {
						param.Parameter.Name = queryFieldName + "." + param.Parameter.Name
						parameters = append(parameters, subParam)
					}
				}
			}
		}
	} else if field.Desc.Kind() != protoreflect.GroupKind {
		// schemaOrReferenceForField also handles array types
		fieldSchema := g.reflect.schemaOrReferenceForField(field, nil)

		parameters = append(parameters,
			&v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:        queryFieldName,
						In:          "query",
						Description: fieldDescription,
						Required:    false,
						Schema:      fieldSchema,
					},
				},
			})
	}

	return parameters
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
func (g *OpenAPIv3Generator) addOperationToDocumentV3(d *v3.Document, op *v3.Operation, path, methodName string) {
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

// addPathsToDocumentV3 adds paths from a specified file descriptor.
func (g *OpenAPIv3Generator) addPathsToDocumentV3(d *v3.Document, services []*protogen.Service) {
	for _, service := range services {
		extService, _ := proto.GetExtension(service.Desc.Options(), E_Service).(*Service)

		annotationsCount := 0
		for _, method := range service.Methods {
			comment := g.filterCommentString(method.Comments.Leading)
			inputMessage := method.Input
			outputMessage := method.Output
			operationID := service.GoName + "_" + method.GoName

			rules := make([]*annotations.HttpRule, 0)

			extHTTP := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
			if extHTTP != nil && extHTTP != annotations.E_Http.InterfaceOf(annotations.E_Http.Zero()) {
				annotationsCount++

				rule := extHTTP.(*annotations.HttpRule)
				rules = append(rules, rule)
				rules = append(rules, rule.AdditionalBindings...)
			}

			for _, rule := range rules {
				var path string
				var methodName string
				var body string

				body = rule.Body
				switch pattern := rule.Pattern.(type) {
				case *annotations.HttpRule_Get:
					path = pattern.Get
					methodName = "GET"
				case *annotations.HttpRule_Post:
					path = pattern.Post
					methodName = "POST"
				case *annotations.HttpRule_Put:
					path = pattern.Put
					methodName = "PUT"
				case *annotations.HttpRule_Delete:
					path = pattern.Delete
					methodName = "DELETE"
				case *annotations.HttpRule_Patch:
					path = pattern.Patch
					methodName = "PATCH"
				case *annotations.HttpRule_Custom:
					path = "custom-unsupported"
				default:
					path = "unknown-unsupported"
				}

				if methodName != "" {
					defaultHost := proto.GetExtension(service.Desc.Options(), annotations.E_DefaultHost).(string)

					op, path2 := g.buildOperationV3(
						d, operationID, service.GoName, comment, defaultHost, path, body, inputMessage, outputMessage)

					// Merge any `Service` annotations with the current
					if extService != nil {
						op.Parameters = append(op.Parameters, extService.Parameters...)
						op.SpecificationExtension = append(op.SpecificationExtension, extService.SpecificationExtension...)
						op.Tags = append(op.Tags, extService.Tags...)
						op.Servers = append(op.Servers, extService.Servers...)
						op.Security = append(op.Security, extService.Security...)
						if extService.ExternalDocs != nil {
							proto.Merge(op.ExternalDocs, extService.ExternalDocs)
						}
					}

					// Merge any `Operation` annotations with the current
					extOperation := proto.GetExtension(method.Desc.Options(), v3.E_Operation)
					if extOperation != nil {
						proto.Merge(op, extOperation.(*v3.Operation))
					}

					for _, v := range op.Parameters {
						if v.Oneof == nil {
							continue
						}

						switch v1 := v.Oneof.(type) {
						case *v3.ParameterOrReference_Parameter:
							p := v1.Parameter
							if p.In == "header" {
								if p.Schema == nil {
									p.Schema = wellknown.NewStringSchema()
								}
							}
						}
					}

					var tags []string
					for _, v := range op.Tags {
						if strings.Contains(v, "=") {
							tagNames := strings.SplitN(v, "=", 2)
							op.SpecificationExtension = append(op.SpecificationExtension, &v3.NamedAny{
								Name: strings.TrimSpace(tagNames[0]),
								Value: &v3.Any{
									Yaml: strings.TrimSpace(tagNames[1]),
								},
							})
							continue
						}

						tags = append(tags, v)
					}

					var extMap = make(map[string]*v3.NamedAny)
					for _, v := range op.SpecificationExtension {
						extMap[v.Name] = v
					}

					op.SpecificationExtension = op.SpecificationExtension[:0]
					for _, v := range extMap {
						op.SpecificationExtension = append(op.SpecificationExtension, v)
					}

					op.Tags = tags
					g.addOperationToDocumentV3(d, op, path2, methodName)
				}
			}
		}

		if annotationsCount > 0 {
			comment := g.filterCommentString(service.Comments.Leading)
			d.Tags = append(d.Tags, &v3.Tag{Name: service.GoName, Description: comment})
		}
	}
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

		schemaName := g.reflect.formatMessageName(message.Desc)

		// Only generate this if we need it and haven't already generated it.
		if !contains(g.reflect.requiredSchemas, schemaName) ||
			contains(g.generatedSchemas, schemaName) {
			continue
		}

		typeName := fullMessageTypeName(message.Desc)
		messageDescription := g.filterCommentString(message.Comments.Leading)

		// `google.protobuf.Value` and `google.protobuf.Any` have special JSON transcoding
		// so we can't just reflect on the message descriptor.
		if typeName == ".google.protobuf.Value" {
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleProtobufValueSchema(schemaName))
			continue
		} else if typeName == ".google.protobuf.Any" {
			g.addSchemaToDocumentV3(d, wellknown.NewGoogleProtobufAnySchema(schemaName))
			continue
		} else if typeName == ".google.rpc.Status" {
			anySchemaName := g.reflect.formatMessageName(anyProtoDesc)
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
			description := g.filterCommentString(field.Comments.Leading)
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
							required = append(required, g.reflect.formatFieldName(field.Desc))
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
					Name:  g.reflect.formatFieldName(field.Desc),
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
