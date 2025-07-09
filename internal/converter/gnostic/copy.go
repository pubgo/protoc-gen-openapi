package gnostic

import (
	goa3 "github.com/google/gnostic/openapiv3"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func ToSecurityRequirements(securityReq []*goa3.SecurityRequirement) []*base.SecurityRequirement {
	return toSecurityRequirements(securityReq)
}

func ToServers(servers []*goa3.Server) []*v3.Server {
	if len(servers) == 0 {
		return nil
	}

	result := make([]*v3.Server, len(servers))
	for i, server := range servers {
		result[i] = ToServer(server)
	}
	return result
}

func ToServer(server *goa3.Server) *v3.Server {
	if server == nil {
		return nil
	}

	return &v3.Server{
		URL:         server.Url,
		Description: server.Description,
		Variables:   toVariables(server.Variables),
	}
}

func ToParametersMap(params *goa3.ParametersOrReferences) *orderedmap.Map[string, *v3.Parameter] {
	m := orderedmap.New[string, *v3.Parameter]()
	for _, item := range params.GetAdditionalProperties() {
		m.Set(item.Name, toParameterV1(item.GetValue()))
	}
	return m
}

func ToParameter(params []*goa3.Parameter) []*v3.Parameter {
	return lo.Map(params, func(item *goa3.Parameter, _ int) *v3.Parameter {
		return toParameterV1(&goa3.ParameterOrReference{Oneof: &goa3.ParameterOrReference_Parameter{Parameter: item}})
	})
}

func ToExtensions(items []*goa3.NamedAny) *orderedmap.Map[string, *yaml.Node] {
	if items == nil {
		return nil
	}

	extensions := orderedmap.New[string, *yaml.Node]()
	for _, namedAny := range items {
		extensions.Set(namedAny.Name, namedAny.Value.ToRawInfo())
	}
	return extensions
}

func toParameterV1(paramOrRef *goa3.ParameterOrReference) *v3.Parameter {
	if paramOrRef == nil || paramOrRef.GetParameter() == nil {
		return nil
	}

	param := paramOrRef.GetParameter()
	p := &v3.Parameter{
		Name:            param.GetName(),
		In:              param.In,
		Description:     param.Description,
		Required:        &param.Required,
		Deprecated:      param.Deprecated,
		AllowEmptyValue: param.AllowEmptyValue,
		Style:           param.Style,
		Explode:         &param.Explode,
		AllowReserved:   param.AllowReserved,
		Schema:          toSchemaOrReference(param.GetSchema()),
		Content:         toMediaTypes(param.GetContent()),
		Extensions:      ToExtensions(param.GetSpecificationExtension()),
	}

	if param.Example != nil {
		p.Example = param.Example.ToRawInfo()
	}
	return p
}
