package generator

import (
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pubgo/protoc-gen-openapi/generator/pure"
)

// buildPathParameters 构建路径参数
func (g *OpenAPIv3Generator) buildPathParameters(path string, inputMessage *protogen.Message) ([]*v3.ParameterOrReference, []string, string) {
	var parameters []*v3.ParameterOrReference
	coveredParameters := make([]string, 0)

	// 处理简单路径参数 {id}
	if allMatches := g.pathPattern.FindAllStringSubmatch(path, -1); allMatches != nil {
		for _, matches := range allMatches {
			coveredParameters = append(coveredParameters, matches[1])
			pathParameter := g.findAndFormatFieldName(matches[1], inputMessage)
			path = strings.Replace(path, matches[1], pathParameter, 1)

			var fieldSchema *v3.SchemaOrReference
			var fieldDescription string
			field := g.findField(pathParameter, inputMessage)
			if field != nil {
				fieldSchema = g.reflect.schemaOrReferenceForField(field, nil)
				fieldDescription = g.filterCommentString(field.Comments.Leading)
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

	// 处理命名路径参数 {name=shelves/*}
	if matches := g.namedPathPattern.FindStringSubmatch(path); matches != nil {
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

// buildServer 构建服务器信息
func (g *OpenAPIv3Generator) buildServer(defaultHost string) []*v3.Server {
	return pure.BuildServer(defaultHost)
}

// buildRequestBody 构建请求体
func (g *OpenAPIv3Generator) buildRequestBody(bodyField string, inputMessage *protogen.Message) *v3.RequestBodyOrReference {
	return pure.BuildRequestBody(bodyField, inputMessage, g.reflect.schemaOrReferenceForMessage)
}

// buildResponses 构建响应
func (g *OpenAPIv3Generator) buildResponses(outputMessage *protogen.Message, defaultResponse bool, doc *v3.Document) *v3.Responses {
	return pure.BuildResponses(
		outputMessage,
		defaultResponse,
		g.reflect.responseContentForMessage,
		g.reflect.formatMessageName,
		g.addSchemaToDocumentV3,
		doc,
		anyProtoDesc,
		statusProtoDesc,
	)
}

// buildOperation 构建操作
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
	return pure.BuildOperation(
		d,
		operationID,
		tagName,
		description,
		defaultHost,
		path,
		bodyField,
		inputMessage,
		outputMessage,
		g.buildPathParameters,
		g.buildQueryParamsV3,
		g.buildServer,
		g.buildRequestBody,
		g.buildResponses,
		*g.conf.DefaultResponse,
	)
}
