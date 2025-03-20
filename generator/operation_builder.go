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
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/pubgo/protoc-gen-openapi/generator/pure"
)

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
