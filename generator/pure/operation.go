// Package pure contains pure functions for OpenAPI v3 generation.
package pure

import (
	"log"
	"net/url"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
)

// BuildServer 构建服务器信息的纯函数版本
func BuildServer(defaultHost string) []*v3.Server {
	if defaultHost == "" {
		return nil
	}

	hostURL, err := url.Parse(defaultHost)
	if err != nil {
		return nil
	}

	hostURL.Scheme = "https"
	return []*v3.Server{{Url: hostURL.String()}}
}

// BuildRequestBody 构建请求体的纯函数版本
func BuildRequestBody(
	bodyField string,
	inputMessage *protogen.Message,
	schemaFunc func(protoreflect.MessageDescriptor) *v3.SchemaOrReference,
) *v3.RequestBodyOrReference {
	if bodyField == "" {
		return nil
	}

	var requestSchema *v3.SchemaOrReference
	if bodyField == "*" {
		requestSchema = schemaFunc(inputMessage.Desc)
	} else {
		for _, field := range inputMessage.Fields {
			if string(field.Desc.Name()) == bodyField {
				switch field.Desc.Kind() {
				case protoreflect.StringKind:
					requestSchema = &v3.SchemaOrReference{
						Oneof: &v3.SchemaOrReference_Schema{
							Schema: &v3.Schema{
								Type: "string",
							},
						},
					}
				case protoreflect.MessageKind:
					requestSchema = schemaFunc(field.Message.Desc)
				default:
					log.Printf("unsupported field type %+v", field.Desc)
				}
				break
			}
		}
	}

	return &v3.RequestBodyOrReference{
		Oneof: &v3.RequestBodyOrReference_RequestBody{
			RequestBody: &v3.RequestBody{
				Required: true,
				Content: &v3.MediaTypes{
					AdditionalProperties: []*v3.NamedMediaType{
						{
							Name: "application/json",
							Value: &v3.MediaType{
								Schema: requestSchema,
							},
						},
					},
				},
			},
		},
	}
}

// BuildResponses 构建响应的纯函数版本
func BuildResponses(
	outputMessage *protogen.Message,
	defaultResponse bool,
	responseFunc func(protoreflect.MessageDescriptor) (string, *v3.MediaTypes),
	formatMessageFunc func(protoreflect.MessageDescriptor) string,
	addSchemaFunc func(*v3.Document, *v3.NamedSchemaOrReference),
	doc *v3.Document,
	anyProtoDesc protoreflect.MessageDescriptor,
	statusProtoDesc protoreflect.MessageDescriptor,
) *v3.Responses {
	name, content := responseFunc(outputMessage.Desc)
	responses := &v3.Responses{
		ResponseOrReference: []*v3.NamedResponseOrReference{
			{
				Name: name,
				Value: &v3.ResponseOrReference{
					Oneof: &v3.ResponseOrReference_Response{
						Response: &v3.Response{
							Description: "OK",
							Content:     content,
						},
					},
				},
			},
		},
	}

	if defaultResponse {
		anySchemaName := formatMessageFunc(anyProtoDesc)
		anySchema := wellknown.NewGoogleProtobufAnySchema(anySchemaName)
		addSchemaFunc(doc, anySchema)

		statusSchemaName := formatMessageFunc(statusProtoDesc)
		statusSchema := wellknown.NewGoogleRpcStatusSchema(statusSchemaName, anySchemaName)
		addSchemaFunc(doc, statusSchema)

		defaultResponse := &v3.NamedResponseOrReference{
			Name: "default",
			Value: &v3.ResponseOrReference{
				Oneof: &v3.ResponseOrReference_Response{
					Response: &v3.Response{
						Description: "Default error response",
						Content: wellknown.NewApplicationJsonMediaType(&v3.SchemaOrReference{
							Oneof: &v3.SchemaOrReference_Reference{
								Reference: &v3.Reference{XRef: "#/components/schemas/" + statusSchemaName},
							},
						}),
					},
				},
			},
		}

		responses.ResponseOrReference = append(responses.ResponseOrReference, defaultResponse)
	}

	return responses
}

// BuildOperation 构建操作的纯函数版本
func BuildOperation(
	doc *v3.Document,
	operationID string,
	tagName string,
	description string,
	defaultHost string,
	path string,
	bodyField string,
	inputMessage *protogen.Message,
	outputMessage *protogen.Message,
	buildPathParamsFunc func(string, *protogen.Message) ([]*v3.ParameterOrReference, []string, string),
	buildQueryParamsFunc func(*protogen.Field) []*v3.ParameterOrReference,
	buildServerFunc func(string) []*v3.Server,
	buildRequestBodyFunc func(string, *protogen.Message) *v3.RequestBodyOrReference,
	buildResponsesFunc func(*protogen.Message, bool, *v3.Document) *v3.Responses,
	defaultResponse bool,
) (*v3.Operation, string) {
	// 构建路径参数
	parameters, coveredParameters, newPath := buildPathParamsFunc(path, inputMessage)

	// 添加请求体参数到已覆盖列表
	if bodyField != "" {
		coveredParameters = append(coveredParameters, bodyField)
	}

	// 添加查询参数
	if bodyField != "*" && string(inputMessage.Desc.FullName()) != "google.api.HttpBody" {
		for _, field := range inputMessage.Fields {
			fieldName := string(field.Desc.Name())
			if !contains(coveredParameters, fieldName) && fieldName != bodyField {
				fieldParams := buildQueryParamsFunc(field)
				parameters = append(parameters, fieldParams...)
			}
		}
	}

	// 创建操作
	op := &v3.Operation{
		Tags:        []string{tagName},
		Description: description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   buildResponsesFunc(outputMessage, defaultResponse, doc),
		Servers:     buildServerFunc(defaultHost),
		RequestBody: buildRequestBodyFunc(bodyField, inputMessage),
	}

	return op, newPath
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
