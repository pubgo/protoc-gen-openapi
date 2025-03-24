package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/pubgo/protoc-gen-openapi/generator"
	"github.com/pubgo/protoc-gen-openapi/internal/converter"
	_ "github.com/pubgo/protoc-gen-openapi/internal/logging"
	"github.com/pubgo/protoc-gen-openapi/version"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	allowGetFlag                = flag.Bool("allow-get", false, "For methods that have `IdempotencyLevel=IDEMPOTENT`, this option will generate HTTP `GET` requests instead of `POST`.")
	baseFlag                    = flag.String("base", "", "The path to a base OpenAPI file to populate fields that this tool doesn't populate.")
	contentTypesFlag            = flag.String("content-types", "json;proto", "Semicolon-separated content types to generate requests/responses")
	debugFlag                   = flag.Bool("debug", false, "Emit debug logs")
	formatFlag                  = flag.String("format", "yaml", "Which format to use for the OpenAPI file, defaults to `yaml`.")
	ignoreGoogleApiHttpFlag     = flag.Bool("ignore-googleapi-http", false, "Ignore `google.api.http` options on methods when generating openapi specs")
	includeNumberEnumValuesFlag = flag.Bool("include-number-enum-values", false, "Include number enum values beside the string versions, defaults to only showing strings")
	pathFlag                    = flag.String("path", "", "Output filepath, defaults to per-protoFile output if not given.")
	pathPrefixFlag              = flag.String("path-prefix", "", "Prefixes the given string to the beginning of each HTTP path.")
	protoFlag                   = flag.Bool("proto", false, "Generate requests/responses with the protobuf content type")
	servicesFlag                = flag.String("services", "", "Filter which services have OpenAPI spec generated. The default is all services. Comma-separated, uses the full path of the service \"[package name].[service name]\"")
	trimUnusedTypesFlag         = flag.Bool("trim-unused-types", false, "Remove types that aren't references from any method request or response.")
	withProtoAnnotationsFlag    = flag.Bool("with-proto-annotations", false, "Add protobuf type annotations to the end of descriptions so users know the protobuf type that the field converts to.")
	withProtoNamesFlag          = flag.Bool("with-proto-names", false, "Use protobuf field names instead of the camelCase JSON names for property names.")
	withStreamingFlag           = flag.Bool("with-streaming", false, "Generate OpenAPI for client/server/bidirectional streaming RPCs (can be messy).")
)

var conf = generator.Configuration{
	Version:         flag.String("version", "0.0.1", "version number text, e.g. 1.2.3"),
	Title:           flag.String("title", "", "name of the API"),
	Description:     flag.String("description", "", "description of the API"),
	Naming:          flag.String("naming", "json", `naming convention. Use "proto" for passing names directly from the proto files`),
	FQSchemaNaming:  flag.Bool("fq_schema_naming", false, `schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name`),
	EnumType:        flag.String("enum_type", "integer", `type for enum serialization. Use "string" for string-based serialization`),
	CircularDepth:   flag.Int("depth", 2, "depth of recursion for circular messages"),
	DefaultResponse: flag.Bool("default_response", true, `add default response. If "true", automatically adds a default response to operations which use the google.rpc.Status message. Useful if you use envoy or grpc-gateway to transcode as they use this type for their default error responses.`),
	OutputMode:      flag.String("output_mode", "merged", `output generation mode. By default, a single openapi.yaml is generated at the out folder. Use "source_relative' to generate a separate '[inputfile].openapi.yaml' next to each '[inputfile].proto'.`),
}

var showVersion = flag.Bool("version", false, "print the version and exit")

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("protoc-gen-openapi %s\n", version.FullVersion())
		return
	}

	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

		}

		return nil
	})

	resp, err := converter.ConvertFrom(os.Stdin)
	if err != nil {
		message := fmt.Sprintf("Failed to read input: %v", err)
		slog.Error(message)
		renderResponse(&pluginpb.CodeGeneratorResponse{
			Error: &message,
		})
		os.Exit(1)
	}

	renderResponse(resp)
}

func renderResponse(resp *pluginpb.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal response", slog.Any("error", err))
		return
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		slog.Error("failed to write response", slog.Any("error", err))
		return
	}
}
