package options

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Config struct {
	AllowGetFlag                   *bool
	BaseFlag                       *string
	ContentTypesFlag               *string
	DebugFlag                      *bool
	FormatFlag                     *string
	IgnoreGoogleApiHttpFlag        *bool
	IncludeNumberEnumValuesFlag    *bool
	PathFlag                       *string
	PathPrefixFlag                 *string
	ProtoFlag                      *bool
	ServicesFlag                   *string
	TrimUnusedTypesFlag            *bool
	WithProtoAnnotationsFlag       *bool
	WithProtoNamesFlag             *bool
	WithStreamingFlag              *bool
	FullyQualifiedMessageNamesFlag *bool
	WithServiceDescriptions        *bool
}

func (c Config) ToOptions() (Options, error) {
	opts := NewOptions()

	opts.Debug = lo.FromPtr(c.DebugFlag)
	opts.IncludeNumberEnumValues = lo.FromPtr(c.IncludeNumberEnumValuesFlag)
	opts.AllowGET = lo.FromPtr(c.AllowGetFlag)
	opts.WithStreaming = lo.FromPtr(c.WithStreamingFlag)
	opts.WithProtoNames = lo.FromPtr(c.WithProtoNamesFlag)
	opts.WithProtoAnnotations = lo.FromPtr(c.WithProtoAnnotationsFlag)
	opts.TrimUnusedTypes = lo.FromPtr(c.TrimUnusedTypesFlag)
	opts.FullyQualifiedMessageNames = lo.FromPtr(c.FullyQualifiedMessageNamesFlag)
	opts.WithServiceDescriptions = lo.FromPtr(c.WithServiceDescriptions)
	opts.IgnoreGoogleapiHTTP = lo.FromPtr(c.IgnoreGoogleApiHttpFlag)
	opts.Path = lo.FromPtr(c.PathFlag)
	opts.PathPrefix = lo.FromPtr(c.PathPrefixFlag)
	opts.Format = lo.FromPtr(c.FormatFlag)
	if !lo.Contains([]string{"yaml", "json"}, opts.Format) {
		return opts, fmt.Errorf("format be yaml or json, not '%s'", opts.Format)
	}

	contentTypes := map[string]struct{}{}
	supportedProtocols := map[string]struct{}{}
	for _, proto := range Protocols {
		supportedProtocols[proto.Name] = struct{}{}
	}
	for _, contentType := range strings.Split(lo.FromPtr(c.ContentTypesFlag), ";") {
		contentType = strings.TrimSpace(contentType)
		_, isSupportedProtocol := supportedProtocols[contentType]
		if !isSupportedProtocol {
			return opts, fmt.Errorf("invalid content type: '%s'", contentType)
		}
		contentTypes[contentType] = struct{}{}
	}

	basePath := lo.FromPtr(c.BaseFlag)
	if basePath != "" {
		ext := path.Ext(basePath)
		switch ext {
		case ".yaml", ".yml", ".json":
			body, err := os.ReadFile(basePath)
			if err != nil {
				return opts, err
			}
			opts.BaseOpenAPI = body
		default:
			return opts, fmt.Errorf("the file extension for 'base' should end with yaml or json, not '%s'", ext)
		}
	}

	for _, service := range strings.Split(lo.FromPtr(c.ServicesFlag), ",") {
		opts.Services = append(opts.Services, protoreflect.FullName(service))
	}

	if len(contentTypes) > 0 {
		opts.ContentTypes = contentTypes
	}
	return opts, nil
}

type Options struct {
	// Format is either 'yaml' or 'json' and is the format of the output OpenAPI file(s).
	Format string

	// BaseOpenAPI is the file contents of a base OpenAPI file.
	BaseOpenAPI []byte

	// WithStreaming will content types related to streaming (warning: can be messy).
	WithStreaming bool

	// AllowGET will let methods with `idempotency_level = NO_SIDE_EFFECTS` to be documented with GET requests.
	AllowGET bool

	// ContentTypes is a map of all content types. Available values are in Protocols.
	ContentTypes map[string]struct{}

	// Debug enables debug logging if set to true.
	Debug bool

	// IncludeNumberEnumValues indicates if numbers are included for enum values in addition to the string representations.
	IncludeNumberEnumValues bool

	// WithProtoNames indicates if protobuf field names should be used instead of JSON names.
	WithProtoNames bool

	// Path is the output OpenAPI path.
	Path string

	// PathPrefix is a prefix that is prepended to every HTTP path.
	PathPrefix string

	// TrimUnusedTypes will remove types that aren't referenced by a service.
	TrimUnusedTypes bool

	// WithProtoAnnotations will add some protobuf annotations for descriptions
	WithProtoAnnotations bool

	// FullyQualifiedMessageNames uses the full path for message types: {pkg}.{name} instead of just the name. This
	// is helpful if you are mixing types from multiple services.
	FullyQualifiedMessageNames bool

	// WithServiceDescriptions set to true will cause service names and their comments to be added to the end of info.description.
	WithServiceDescriptions bool

	// IgnoreGoogleapiHTTP set to true will cause service to always generate OpenAPI specs for connect endpoints, and ignore any google.api.http options.
	IgnoreGoogleapiHTTP bool

	// Services filters which services will be used for generating OpenAPI spec.
	Services []protoreflect.FullName

	MessageAnnotator        MessageAnnotator
	FieldAnnotator          FieldAnnotator
	FieldReferenceAnnotator FieldReferenceAnnotator
}

func (opts Options) HasService(serviceName protoreflect.FullName) bool {
	if len(opts.Services) == 0 {
		return true
	}
	for _, service := range opts.Services {
		if service == serviceName {
			return true
		}
	}
	return false
}

func NewOptions() Options {
	return Options{
		Format: "yaml",
		ContentTypes: map[string]struct{}{
			"json": {},
		},
	}
}

func IsValidContentType(contentType string) bool {
	for _, protocol := range Protocols {
		if protocol.Name == contentType {
			return true
		}
	}
	return false
}
