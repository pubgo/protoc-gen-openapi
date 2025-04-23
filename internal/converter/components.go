package converter

import (
	"log/slog"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	highv3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/util"
)

func fileToComponents(opts options.Options, fd protoreflect.FileDescriptor) (*highv3.Components, error) {
	// Add schema from messages/enums
	components := &highv3.Components{
		Schemas:         orderedmap.New[string, *base.SchemaProxy](),
		Responses:       orderedmap.New[string, *v3.Response](),
		Parameters:      orderedmap.New[string, *v3.Parameter](),
		Examples:        orderedmap.New[string, *base.Example](),
		RequestBodies:   orderedmap.New[string, *v3.RequestBody](),
		Headers:         orderedmap.New[string, *v3.Header](),
		SecuritySchemes: orderedmap.New[string, *v3.SecurityScheme](),
		Links:           orderedmap.New[string, *v3.Link](),
		Callbacks:       orderedmap.New[string, *v3.Callback](),
		Extensions:      orderedmap.New[string, *yaml.Node](),
	}
	st := NewState(opts)
	slog.Debug("start collection")
	st.CollectFile(fd)
	slog.Debug("collection complete", slog.String("file", string(fd.Name())), slog.Int("messages", len(st.Messages)), slog.Int("enum", len(st.Enums)))
	components.Schemas = stateToSchema(st)

	hasGetRequests := false
	hasMethods := false

	// Add requestBodies and responses for methods
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			hasGet := methodHasGet(opts, method)
			if hasGet {
				hasGetRequests = true
			}
			hasMethods = true
		}
	}

	if hasGetRequests {
		components.Schemas.Set("encoding", base.CreateSchemaProxy(&base.Schema{
			Title:       "encoding",
			Description: "Define which encoding or 'Message-Codec' to use",
			Enum: []*yaml.Node{
				utils.CreateStringNode("proto"),
				utils.CreateStringNode("json"),
			},
		}))

		components.Schemas.Set("base64", base.CreateSchemaProxy(&base.Schema{
			Title:       "base64",
			Description: "Specifies if the message query param is base64 encoded, which may be required for binary data",
			Type:        []string{"boolean"},
		}))

		components.Schemas.Set("compression", base.CreateSchemaProxy(&base.Schema{
			Title:       "compression",
			Description: "Which compression algorithm to use for this request",
			Enum: []*yaml.Node{
				utils.CreateStringNode("identity"),
				utils.CreateStringNode("gzip"),
				utils.CreateStringNode("br"),
			},
		}))
		components.Schemas.Set("lava", base.CreateSchemaProxy(&base.Schema{
			Title:       "lava",
			Description: "Define the version of the Lava protocol",
			Enum: []*yaml.Node{
				utils.CreateStringNode("v1"),
			},
		}))
	}

	if hasMethods {
		components.Schemas.Set("lava-protocol-version", base.CreateSchemaProxy(&base.Schema{
			Title:       "Lava-Protocol-Version",
			Description: "Define the version of the Lava protocol",
			Type:        []string{"number"},
			Enum:        []*yaml.Node{utils.CreateIntNode("1")},
			Const:       utils.CreateIntNode("1"),
		}))

		components.Schemas.Set("lava-timeout-header", base.CreateSchemaProxy(&base.Schema{
			Title:       "Lava-Timeout-Ms",
			Description: "Define the timeout, in ms",
			Type:        []string{"number"},
		}))

		getErrorProps := func() *orderedmap.Map[string, *base.SchemaProxy] {
			connectErrorProps := orderedmap.New[string, *base.SchemaProxy]()
			connectErrorProps.Set("status_code", base.CreateSchemaProxy(&base.Schema{
				Title:       "status code",
				Description: "GRPC code corresponding to HTTP status code, which can be converted to each other",
				Type:        []string{"string"},
				Format:      "enum",
				Examples:    []*yaml.Node{utils.CreateStringNode("OK")},
				Enum: []*yaml.Node{
					utils.CreateStringNode("OK"),
					utils.CreateStringNode("Canceled"),
					utils.CreateStringNode("InvalidArgument"),
					utils.CreateStringNode("DeadlineExceeded"),
					utils.CreateStringNode("NotFound"),
					utils.CreateStringNode("AlreadyExists"),
					utils.CreateStringNode("PermissionDenied"),
					utils.CreateStringNode("ResourceExhausted"),
					utils.CreateStringNode("FailedPrecondition"),
					utils.CreateStringNode("Aborted"),
					utils.CreateStringNode("OutOfRange"),
					utils.CreateStringNode("Unimplemented"),
					utils.CreateStringNode("Internal"),
					utils.CreateStringNode("Unavailable"),
					utils.CreateStringNode("DataLoss"),
					utils.CreateStringNode("Unauthenticated"),
				},
			}))
			connectErrorProps.Set("name", base.CreateSchemaProxy(&base.Schema{
				Description: "Error name, e.g. lava.auth.token_not_found.",
				Type:        []string{"string"},
			}))
			connectErrorProps.Set("message", base.CreateSchemaProxy(&base.Schema{
				Description: "Error message, e.g. token not found",
				Type:        []string{"string"},
			}))
			connectErrorProps.Set("code", base.CreateSchemaProxy(&base.Schema{
				Description: "Business Code, e.g. 200001",
				Type:        []string{"number"},
			}))
			connectErrorProps.Set("details", base.CreateSchemaProxy(&base.Schema{
				Title:       "details",
				Description: "Error detail include request or other user defined information",
				Type:        []string{"array"},
				Items:       &base.DynamicValue[*base.SchemaProxy, bool]{A: base.CreateSchemaProxyRef("#/components/schemas/google.protobuf.Any")},
			}))
			return connectErrorProps
		}

		components.Schemas.Set("lava.error", base.CreateSchemaProxy(&base.Schema{
			Title:                "Lava Error",
			Description:          `Error type returned by lava: https://github.com/pubgo/funk/blob/master/proto/errorpb/errors.proto`,
			Properties:           getErrorProps(),
			Type:                 []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{N: 1, B: true},
		}))
		anyPair := util.NewGoogleAny()
		components.Schemas.Set(anyPair.ID, base.CreateSchemaProxy(anyPair.Schema))
	}

	return components, nil
}
