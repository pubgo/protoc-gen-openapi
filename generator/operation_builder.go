/*
 * @Author: barry
 * @Date: 2025-03-20 10:10:01
 * @LastEditors: barry
 * @LastEditTime: 2025-03-20 10:10:29
 * @Description:
 */
package generator

import (
	v3 "github.com/google/gnostic-models/openapiv3"
	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"log"
	"net/url"
)

// buildServer 构建服务器信息
func (g *OpenAPIv3Generator) buildServer(defaultHost string) []*v3.Server {
	return BuildServer(defaultHost)
}

// buildRequestBody 构建请求体
func (g *OpenAPIv3Generator) buildRequestBody(bodyField string, inputMessage *protogen.Message) *v3.RequestBodyOrReference {
	return BuildRequestBody(bodyField, inputMessage, g.reflect.schemaOrReferenceForMessage)
}

// buildResponses 构建响应
func (g *OpenAPIv3Generator) buildResponses(outputMessage *protogen.Message, defaultResponse bool, doc *v3.Document) *v3.Responses {
	return BuildResponses(
		outputMessage,
		defaultResponse,
		g.reflect.responseContentForMessage,
		g.addSchemaToDocumentV3,
		doc,
		anyProtoDesc,
		statusProtoDesc,
		&g.conf,
	)
}

func BuildResponses(
	outputMessage *protogen.Message,
	defaultResponse bool,
	responseFunc func(protoreflect.MessageDescriptor) (string, *v3.MediaTypes),
	addSchemaFunc func(*v3.Document, *v3.NamedSchemaOrReference),
	doc *v3.Document,
	anyProtoDesc protoreflect.MessageDescriptor,
	statusProtoDesc protoreflect.MessageDescriptor,
	cfg *Configuration,
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
		anySchemaName := formatMessageName(cfg, anyProtoDesc)
		anySchema := wellknown.NewGoogleProtobufAnySchema(anySchemaName)
		addSchemaFunc(doc, anySchema)

		statusSchemaName := formatMessageName(cfg, statusProtoDesc)
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
