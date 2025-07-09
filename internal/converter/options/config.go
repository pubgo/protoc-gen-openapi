package options

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pubgo/funk"
	"github.com/pubgo/funk/assert"
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

	supportedProtocolMap := lo.SliceToMap(Protocols, func(proto Protocol) (string, Protocol) { return proto.Name, proto })
	opts.ContentTypes = lo.SliceToMap(strings.Split(lo.FromPtr(c.ContentTypesFlag), ";"), func(contentType string) (string, struct{}) {
		contentType = strings.TrimSpace(contentType)
		_, isSupportedProtocol := supportedProtocolMap[contentType]
		if !isSupportedProtocol {
			panic(fmt.Errorf("invalid content type: '%s'", contentType))
		}
		return contentType, struct{}{}
	})

	opts.BaseOpenAPI = funk.DoFunc(func() []byte {
		basePath := lo.FromPtr(c.BaseFlag)
		if basePath == "" {
			return nil
		}

		ext := strings.TrimSpace(path.Ext(basePath))
		if lo.Contains([]string{".yaml", ".yml", ".json"}, ext) {
			return assert.Must1(os.ReadFile(basePath))
		}

		assert.Must(fmt.Errorf("the file extension for 'base' should end with yaml or json, not '%s'", ext))
		return nil
	})

	opts.Services = lo.Map(strings.Split(lo.FromPtr(c.ServicesFlag), ","), func(item string, index int) protoreflect.FullName {
		return protoreflect.FullName(strings.TrimSpace(item))
	})
	opts.Services = lo.Filter(lo.Uniq(opts.Services), func(item protoreflect.FullName, index int) bool { return string(item) != "" })

	return opts, nil
}
