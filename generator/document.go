package generator

import (
	"sort"
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// BuildDocumentV3 构建 OpenAPI v3 文档
func BuildDocumentV3(config Configuration, files []*protogen.File) *v3.Document {
	doc := &v3.Document{
		Openapi: "3.0.3",
		Info: &v3.Info{
			Version:     *config.Version,
			Title:       *config.Title,
			Description: *config.Description,
		},
		Paths: &v3.Paths{
			Path: []*v3.NamedPathItem{},
		},
		Components: &v3.Components{
			Schemas: &v3.SchemasOrReferences{
				AdditionalProperties: []*v3.NamedSchemaOrReference{},
			},
		},
	}

	// 添加服务路径
	for _, file := range files {
		if !file.Generate {
			continue
		}
		AddPathsToDocument(doc, file.Services)
	}

	// 添加 Schema
	for _, file := range files {
		AddSchemasToDocument(doc, file.Messages)
	}

	// 排序标签
	sortTags(doc.Tags)

	// 排序路径
	sortPaths(doc.Paths.Path)

	// 排序 Schema
	sortSchemas(doc.Components.Schemas.AdditionalProperties)

	return doc
}

// AddPathsToDocument 添加服务路径到文档
func AddPathsToDocument(doc *v3.Document, services []*protogen.Service) {
	for _, service := range services {
		for _, method := range service.Methods {
			comment := filterCommentString(method.Comments.Leading)
			inputMessage := method.Input
			outputMessage := method.Output
			operationID := service.GoName + "_" + method.GoName

			rules := make([]*annotations.HttpRule, 0)

			extHTTP := proto.GetExtension(method.Desc.Options(), annotations.E_Http)
			if extHTTP != nil && extHTTP != annotations.E_Http.InterfaceOf(annotations.E_Http.Zero()) {
				rule := extHTTP.(*annotations.HttpRule)
				rules = append(rules, rule)
				rules = append(rules, rule.AdditionalBindings...)
			}

			for _, rule := range rules {
				path, methodName, body := parseHTTPRule(rule)
				if methodName != "" {
					defaultHost := proto.GetExtension(service.Desc.Options(), annotations.E_DefaultHost).(string)
					operation := buildOperation(operationID, service.GoName, comment, defaultHost, path, body, inputMessage, outputMessage)
					addOperationToPath(doc.Paths, path, methodName, operation)
				}
			}
		}
	}
}

// parseHTTPRule 解析 HTTP 规则
func parseHTTPRule(rule *annotations.HttpRule) (path, methodName, body string) {
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
	return
}

// buildOperation 构建操作
func buildOperation(operationID, tagName, comment, defaultHost, path, body string, inputMessage, outputMessage *protogen.Message) *v3.Operation {
	op := &v3.Operation{
		OperationId: operationID,
		Description: comment,
		Tags:        []string{tagName},
	}

	// 添加服务器信息
	if defaultHost != "" {
		op.Servers = []*v3.Server{{
			Url: "https://" + defaultHost,
		}}
	}

	// 构建请求体
	if body != "" && body != "*" {
		op.RequestBody = &v3.RequestBodyOrReference{
			Oneof: &v3.RequestBodyOrReference_RequestBody{
				RequestBody: &v3.RequestBody{
					Required: true,
					Content: &v3.MediaTypes{
						AdditionalProperties: []*v3.NamedMediaType{
							{
								Name: "application/json",
								Value: &v3.MediaType{
									Schema: &v3.SchemaOrReference{
										Oneof: &v3.SchemaOrReference_Reference{
											Reference: &v3.Reference{
												XRef: schemaReferenceForMessage(inputMessage.Desc),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	// 构建路径参数
	pathParams := extractPathParams(path)
	if len(pathParams) > 0 {
		op.Parameters = make([]*v3.ParameterOrReference, 0, len(pathParams))
		for _, param := range pathParams {
			op.Parameters = append(op.Parameters, &v3.ParameterOrReference{
				Oneof: &v3.ParameterOrReference_Parameter{
					Parameter: &v3.Parameter{
						Name:     param,
						In:       "path",
						Required: true,
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

	// 构建响应
	op.Responses = &v3.Responses{
		ResponseOrReference: []*v3.NamedResponseOrReference{
			{
				Name: "200",
				Value: &v3.ResponseOrReference{
					Oneof: &v3.ResponseOrReference_Response{
						Response: &v3.Response{
							Description: "A successful response.",
							Content: &v3.MediaTypes{
								AdditionalProperties: []*v3.NamedMediaType{
									{
										Name: "application/json",
										Value: &v3.MediaType{
											Schema: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Reference{
													Reference: &v3.Reference{
														XRef: schemaReferenceForMessage(outputMessage.Desc),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "default",
				Value: &v3.ResponseOrReference{
					Oneof: &v3.ResponseOrReference_Response{
						Response: &v3.Response{
							Description: "An unexpected error response.",
							Content: &v3.MediaTypes{
								AdditionalProperties: []*v3.NamedMediaType{
									{
										Name: "application/json",
										Value: &v3.MediaType{
											Schema: &v3.SchemaOrReference{
												Oneof: &v3.SchemaOrReference_Schema{
													Schema: &v3.Schema{
														Type: "object",
														Properties: &v3.Properties{
															AdditionalProperties: []*v3.NamedSchemaOrReference{
																{
																	Name: "code",
																	Value: &v3.SchemaOrReference{
																		Oneof: &v3.SchemaOrReference_Schema{
																			Schema: &v3.Schema{
																				Type:   "integer",
																				Format: "int32",
																			},
																		},
																	},
																},
																{
																	Name: "message",
																	Value: &v3.SchemaOrReference{
																		Oneof: &v3.SchemaOrReference_Schema{
																			Schema: &v3.Schema{
																				Type: "string",
																			},
																		},
																	},
																},
																{
																	Name: "details",
																	Value: &v3.SchemaOrReference{
																		Oneof: &v3.SchemaOrReference_Schema{
																			Schema: &v3.Schema{
																				Type: "array",
																				Items: &v3.ItemsItem{
																					SchemaOrReference: []*v3.SchemaOrReference{
																						{
																							Oneof: &v3.SchemaOrReference_Schema{
																								Schema: &v3.Schema{
																									Type: "object",
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return op
}

// extractPathParams 从路径中提取参数名
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// 处理 {param=*} 格式
			param := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			if idx := strings.Index(param, "="); idx != -1 {
				param = param[:idx]
			}
			params = append(params, param)
		}
	}
	return params
}

// addOperationToPath 添加操作到路径
func addOperationToPath(paths *v3.Paths, path, methodName string, operation *v3.Operation) {
	var selectedPathItem *v3.NamedPathItem
	for _, namedPathItem := range paths.Path {
		if namedPathItem.Name == path {
			selectedPathItem = namedPathItem
			break
		}
	}
	// If we get here, we need to create a path item.
	if selectedPathItem == nil {
		selectedPathItem = &v3.NamedPathItem{Name: path, Value: &v3.PathItem{}}
		paths.Path = append(paths.Path, selectedPathItem)
	}
	// Set the operation on the specified method.
	switch methodName {
	case "GET":
		selectedPathItem.Value.Get = operation
	case "POST":
		selectedPathItem.Value.Post = operation
	case "PUT":
		selectedPathItem.Value.Put = operation
	case "DELETE":
		selectedPathItem.Value.Delete = operation
	case "PATCH":
		selectedPathItem.Value.Patch = operation
	}
}

// AddSchemasToDocument 添加消息 Schema 到文档
func AddSchemasToDocument(doc *v3.Document, messages []*protogen.Message) {
	for _, message := range messages {
		// 递归处理嵌套消息
		AddSchemasToDocument(doc, message.Messages)

		schemaName := formatMessageName(message.Desc)
		schema := buildSchema(message)
		addSchemaToComponents(doc.Components, schemaName, schema)
	}
}

// buildSchema 构建 Schema
func buildSchema(message *protogen.Message) *v3.SchemaOrReference {
	typeName := fullMessageTypeName(message.Desc)
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
		return nil

	case ".google.protobuf.BoolValue":
		return wellknown.NewBooleanSchema()

	case ".google.protobuf.BytesValue":
		return wellknown.NewBytesSchema()

	case ".google.protobuf.Int32Value", ".google.protobuf.UInt32Value":
		return wellknown.NewIntegerSchema(getValueKind(message.Desc))

	case ".google.protobuf.StringValue", ".google.protobuf.Int64Value", ".google.protobuf.UInt64Value":
		return wellknown.NewStringSchema()

	case ".google.protobuf.FloatValue", ".google.protobuf.DoubleValue":
		return wellknown.NewNumberSchema(getValueKind(message.Desc))

	default:
		ref := schemaReferenceForMessage(message.Desc)
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Reference{Reference: &v3.Reference{XRef: ref}},
		}
	}
}

// addSchemaToComponents 添加 Schema 到组件
func addSchemaToComponents(components *v3.Components, name string, schema *v3.SchemaOrReference) {
	if schema == nil {
		return
	}

	// 检查是否已存在
	for _, existing := range components.Schemas.AdditionalProperties {
		if existing.Name == name {
			return
		}
	}

	// 添加新的 Schema
	components.Schemas.AdditionalProperties = append(
		components.Schemas.AdditionalProperties,
		&v3.NamedSchemaOrReference{
			Name:  name,
			Value: schema,
		},
	)
}

// 排序函数
func sortTags(tags []*v3.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})
}

func sortPaths(paths []*v3.NamedPathItem) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Name < paths[j].Name
	})
}

func sortSchemas(schemas []*v3.NamedSchemaOrReference) {
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})
}
