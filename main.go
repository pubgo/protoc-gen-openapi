package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/pubgo/protoc-gen-openapi/internal/converter"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	_ "github.com/pubgo/protoc-gen-openapi/internal/logging"
	"github.com/pubgo/protoc-gen-openapi/version"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var conf = options.Config{
	AllowGetFlag:                   flag.Bool("allow-get", false, "For methods that have `IdempotencyLevel=IDEMPOTENT`, this option will generate HTTP `GET` requests instead of `POST`."),
	BaseFlag:                       flag.String("base", "", "The path to a base OpenAPI file to populate fields that this tool doesn't populate."),
	ContentTypesFlag:               flag.String("content-types", "json;proto", "Semicolon-separated content types to generate requests/responses"),
	DebugFlag:                      flag.Bool("debug", false, "Emit debug logs"),
	FormatFlag:                     flag.String("format", "yaml", "Which format to use for the OpenAPI file, defaults to `yaml`."),
	IgnoreGoogleApiHttpFlag:        flag.Bool("ignore-googleapi-http", false, "Ignore `google.api.http` options on methods when generating openapi specs"),
	IncludeNumberEnumValuesFlag:    flag.Bool("include-number-enum-values", false, "Include number enum values beside the string versions, defaults to only showing strings"),
	PathFlag:                       flag.String("path", "", "Output filepath, defaults to per-protoFile output if not given."),
	PathPrefixFlag:                 flag.String("path-prefix", "", "Prefixes the given string to the beginning of each HTTP path."),
	ProtoFlag:                      flag.Bool("proto", false, "Generate requests/responses with the protobuf content type"),
	ServicesFlag:                   flag.String("services", "", "Filter which services have OpenAPI spec generated. The default is all services. Comma-separated, uses the full path of the service \"[package name].[service name]\""),
	TrimUnusedTypesFlag:            flag.Bool("trim-unused-types", false, "Remove types that aren't references from any method request or response."),
	WithProtoAnnotationsFlag:       flag.Bool("with-proto-annotations", false, "Add protobuf type annotations to the end of descriptions so users know the protobuf type that the field converts to."),
	WithProtoNamesFlag:             flag.Bool("with-proto-names", false, "Use protobuf field names instead of the camelCase JSON names for property names."),
	WithStreamingFlag:              flag.Bool("with-streaming", false, "Generate OpenAPI for client/server/bidirectional streaming RPCs (can be messy)."),
	FullyQualifiedMessageNamesFlag: flag.Bool("fully-qualified-message-names", false, "Use the full path for message types: {pkg}.{name} instead of just the name. This is helpful if you are mixing types from multiple services."),
	WithServiceDescriptions:        flag.Bool("with-service-descriptions", false, "set to true will cause service names and their comments to be added to the end of info.description"),
}

//var conf = generator.Configuration{
//	Version:         flag.String("version", "0.0.1", "version number text, e.g. 1.2.3"),
//	Title:           flag.String("title", "", "name of the API"),
//	Description:     flag.String("description", "", "description of the API"),
//	Naming:          flag.String("naming", "json", `naming convention. Use "proto" for passing names directly from the proto files`),
//	FQSchemaNaming:  flag.Bool("fq_schema_naming", false, `schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name`),
//	EnumType:        flag.String("enum_type", "integer", `type for enum serialization. Use "string" for string-based serialization`),
//	CircularDepth:   flag.Int("depth", 2, "depth of recursion for circular messages"),
//	DefaultResponse: flag.Bool("default_response", true, `add default response. If "true", automatically adds a default response to operations which use the google.rpc.Status message. Useful if you use envoy or grpc-gateway to transcode as they use this type for their default error responses.`),
//	OutputMode:      flag.String("output_mode", "merged", `output generation mode. By default, a single openapi.yaml is generated at the out folder. Use "source_relative' to generate a separate '[inputfile].openapi.yaml' next to each '[inputfile].proto'.`),
//}

var showVersion = flag.Bool("version", false, "print the version and exit")

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("protoc-gen-openapi %s\n", version.FullVersion())
		return
	}

	var flagSet = func(name, value string) error {
		err := flag.CommandLine.Set(name, value)
		if err != nil {
			slog.Info("openapi flags set error", name, value, "err", err)
		}

		return nil
	}
	var runPlugin = func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2024

		return converter.Convert(gen, conf)
	}
	protogen.Options{ParamFunc: flagSet}.Run(runPlugin)
}
