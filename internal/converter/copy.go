package converter

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	openapiv3 "github.com/google/gnostic-models/openapiv3"
	"github.com/lmittmann/tint"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protoc-gen-openapi/generator"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/gnostic"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/googleapi"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/util"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
)

func ConvertV1(gen *protogen.Plugin, cfg options.Config) error {
	defer recovery.Recovery(func(err error) {
		slog.Error(err.Error())
	})

	opts := assert.Must1(cfg.ToOptions())

	var req = gen.Request
	annotate := &annotator{}
	if opts.MessageAnnotator == nil {
		opts.MessageAnnotator = annotate
	}
	if opts.FieldAnnotator == nil {
		opts.FieldAnnotator = annotate
	}
	if opts.FieldReferenceAnnotator == nil {
		opts.FieldReferenceAnnotator = annotate
	}

	if opts.Debug {
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelDebug})))
	}

	genFiles := lo.SliceToMap(req.FileToGenerate, func(item string) (string, struct{}) { return item, struct{}{} })

	// We need this to resolve dependencies when making protodesc versions of the files
	resolver := assert.Must1(protodesc.NewFiles(&descriptorpb.FileDescriptorSet{
		File: req.GetProtoFile(),
	}))

	newSpec := func() (*v3.Document, error) {
		model := &v3.Document{}
		initializeDoc(model)
		return model, nil
	}
	if len(opts.BaseOpenAPI) > 0 {
		newSpec = func() (*v3.Document, error) {
			document, err := libopenapi.NewDocument(opts.BaseOpenAPI)
			if err != nil {
				return &v3.Document{}, fmt.Errorf("unmarshalling base: %w", err)
			}

			v3Document, errs := document.BuildV3Model()
			if len(errs) > 0 {
				return &v3.Document{}, errors.Join(errs...)
			}
			model := &v3Document.Model
			initializeDoc(model)
			return model, nil
		}
	}

	spec := assert.Must1(newSpec())

	outFiles := map[string]*v3.Document{}
	for _, fileDesc := range req.GetProtoFile() {
		if _, ok := genFiles[fileDesc.GetName()]; !ok {
			continue
		}

		slog.Debug("generating file", slog.String("name", fileDesc.GetName()))

		fd, err := resolver.FindFileByPath(fileDesc.GetName())
		if err != nil {
			slog.Error("error loading file", slog.Any("error", err))
			return err
		}

		// Create a per-file openapi spec if we're not merging all into one
		if opts.Path == "" {
			spec = assert.Must1(newSpec())
			spec.Info.Title = string(fd.FullName())
			spec.Info.Description = util.FormatComments(fd.SourceLocations().ByDescriptor(fd))
		}

		assert.Must(appendToSpec(opts, spec, fd))

		if opts.Path == "" {
			name := fileDesc.GetName()
			filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".openapi." + opts.Format
			outFiles[filename] = spec
		}

		spec.Tags = mergeTags(spec.Tags)
	}

	if opts.Path != "" {
		outFiles[opts.Path] = spec
	}

	for path, doc := range outFiles {
		content := assert.Must1(specToFile(opts, doc))

		gg := gen.NewGeneratedFile(path, "")
		assert.Must1(gg.Write([]byte(content)))
	}

	return nil
}

func mergeOperationV2(existing *v3.Operation, srv *generator.Service) {
	if srv == nil || existing == nil {
		return
	}

	existing.Tags = lo.Uniq(append(existing.Tags, srv.Tags...))
	securityList := lo.Map(srv.Security, func(item *openapiv3.NamedStringArray, index int) *openapiv3.SecurityRequirement {
		return &openapiv3.SecurityRequirement{AdditionalProperties: []*openapiv3.NamedStringArray{item}}
	})
	existing.Security = append(existing.Security, gnostic.ToSecurityRequirements(securityList)...)
	existing.Servers = append(existing.Servers, gnostic.ToServers(srv.Servers)...)
	existing.Parameters = append(existing.Parameters, gnostic.ToParameter(srv.Parameters)...)
	existing.Parameters = lo.UniqBy(existing.Parameters, func(item *v3.Parameter) string { return item.Name + item.In })

	ext := gnostic.ToExtensions(srv.Extensions)
	if ext == nil {
		ext = orderedmap.New[string, *yaml.Node]()
	}

	if existing.Extensions == nil {
		existing.Extensions = ext
	} else {
		for pair := ext.First(); pair != nil; pair = pair.Next() {
			existing.Extensions.Set(pair.Key(), pair.Value())
		}
	}
}

func mergePathItemsV1(existing, new *v3.PathItem) *v3.PathItem {
	// Merge operations
	operations := []struct {
		existingOp **v3.Operation
		newOp      *v3.Operation
	}{
		{&existing.Get, new.Get},
		{&existing.Post, new.Post},
		{&existing.Put, new.Put},
		{&existing.Delete, new.Delete},
		{&existing.Options, new.Options},
		{&existing.Head, new.Head},
		{&existing.Patch, new.Patch},
		{&existing.Trace, new.Trace},
	}

	for _, op := range operations {
		if op.newOp != nil {
			mergeOperation(op.existingOp, op.newOp)
		}
	}

	// Merge other fields
	if new.Summary != "" {
		existing.Summary = new.Summary
	}
	if new.Description != "" {
		existing.Description = new.Description
	}
	existing.Servers = append(existing.Servers, new.Servers...)
	existing.Parameters = append(existing.Parameters, new.Parameters...)

	// Merge extensions
	for pair := new.Extensions.First(); pair != nil; pair = pair.Next() {
		if _, ok := existing.Extensions.Get(pair.Key()); !ok {
			existing.Extensions.Set(pair.Key(), pair.Value())
		}
	}
	return existing
}

var _ = addPathItemsFromFile

func addPathItemsFromFileV1(opts options.Options, fd protoreflect.FileDescriptor, paths *v3.Paths) error {
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}

		srv := googleapi.GetSrvOptions(opts, service)
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			pathItems := googleapi.MakePathItems(opts, method)

			// Helper function to update or set path items
			addPathItem := func(path string, newItem *v3.PathItem) {
				path = util.MakePath(opts, path)
				if existing, ok := paths.PathItems.Get(path); ok {
					newItem = mergePathItemsV1(existing, newItem)
				}

				mergeOperationV2(newItem.Get, srv)
				mergeOperationV2(newItem.Put, srv)
				mergeOperationV2(newItem.Post, srv)
				mergeOperationV2(newItem.Delete, srv)
				mergeOperationV2(newItem.Patch, srv)
				paths.PathItems.Set(path, newItem)
			}

			// Update path items from google.api annotations
			for pair := pathItems.First(); pair != nil; pair = pair.Next() {
				item := gnostic.PathItemWithMethodAnnotations(pair.Value(), method)
				addPathItem(pair.Key(), item)
			}

			// Default to ConnectRPC/gRPC path if no google.api annotations
			if pathItems == nil || pathItems.Len() == 0 {
				path := "/" + string(service.FullName()) + "/" + string(method.Name())
				addPathItem(path, methodToPathItem(opts, method))
			}
		}
	}

	return nil
}

func methodToOperation(opts options.Options, method protoreflect.MethodDescriptor, returnGet bool) *v3.Operation {
	fd := method.ParentFile()
	service := method.Parent().(protoreflect.ServiceDescriptor)
	loc := fd.SourceLocations().ByDescriptor(method)
	op := &v3.Operation{
		Summary:     string(method.Name()),
		OperationId: string(method.FullName()),
		Deprecated:  util.IsMethodDeprecated(method),
		Tags:        []string{string(service.FullName())},
		Description: util.FormatComments(loc),
	}

	isStreaming := method.IsStreamingClient() || method.IsStreamingServer()
	if isStreaming && !opts.WithStreaming {
		return nil
	}

	// Responses
	codeMap := orderedmap.New[string, *v3.Response]()
	outputId := util.FormatTypeRef(string(method.Output().FullName()))
	codeMap.Set("200", &v3.Response{
		Description: "Success",
		Content: util.MakeMediaTypes(
			opts,
			base.CreateSchemaProxyRef("#/components/schemas/"+outputId),
			false,
			isStreaming,
		),
	})
	op.Responses = &v3.Responses{
		Codes: codeMap,
		Default: &v3.Response{
			Description: "Error",
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/lava.error"),
				false,
				isStreaming,
			),
		},
	}

	op.Parameters = append(op.Parameters,
		&v3.Parameter{
			Name:     "Connect-Protocol-Version",
			In:       "header",
			Required: util.BoolPtr(true),
			Schema:   base.CreateSchemaProxyRef("#/components/schemas/connect-protocol-version"),
		},
		&v3.Parameter{
			Name:   "Connect-Timeout-Ms",
			In:     "header",
			Schema: base.CreateSchemaProxyRef("#/components/schemas/connect-timeout-header"),
		},
	)

	// Request parameters
	inputId := util.FormatTypeRef(string(method.Input().FullName()))
	if returnGet {
		op.OperationId = op.OperationId + ".get"
		op.Parameters = append(op.Parameters,
			&v3.Parameter{
				Name: "message",
				In:   "query",
				Content: util.MakeMediaTypes(
					opts,
					base.CreateSchemaProxyRef("#/components/schemas/"+util.FormatTypeRef(inputId)),
					true,
					isStreaming),
			},
			&v3.Parameter{
				Name:     "encoding",
				In:       "query",
				Required: util.BoolPtr(true),
				Schema:   base.CreateSchemaProxyRef("#/components/schemas/encoding"),
			},
			&v3.Parameter{
				Name:   "base64",
				In:     "query",
				Schema: base.CreateSchemaProxyRef("#/components/schemas/base64"),
			},
			&v3.Parameter{
				Name:   "compression",
				In:     "query",
				Schema: base.CreateSchemaProxyRef("#/components/schemas/compression"),
			},
			&v3.Parameter{
				Name:   "connect",
				In:     "query",
				Schema: base.CreateSchemaProxyRef("#/components/schemas/connect"),
			},
		)
	} else {
		op.RequestBody = &v3.RequestBody{
			Content: util.MakeMediaTypes(
				opts,
				base.CreateSchemaProxyRef("#/components/schemas/"+inputId),
				true,
				isStreaming,
			),
			Required: util.BoolPtr(true),
		}
	}

	return op
}

func setComponents(hasMethods bool, components *v3.Components) {
	if !hasMethods {
		return
	}

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

func enumToSchemaV1(state *State, tt protoreflect.EnumDescriptor) (string, *base.Schema) {
	slog.Debug("enumToSchema", slog.Any("descriptor", tt.FullName()))
	children := make([]*yaml.Node, 0)
	values := tt.Values()
	desc := util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt))
	for i := 0; i < values.Len(); i++ {
		value := values.Get(i)
		comment := util.FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(value))
		if comment != "" {
			desc += fmt.Sprintf("- %d, %s: %s\n", value.Number(), value.Name(), comment)
		} else {
			desc += fmt.Sprintf("- %d, %s\n", value.Number(), value.Name())
		}

		children = append(children, utils.CreateStringNode(string(value.Name())))
		if state.Opts.IncludeNumberEnumValues {
			children = append(children, utils.CreateIntNode(strconv.FormatInt(int64(value.Number()), 10)))
		}
	}

	title := string(tt.Name())
	if state.Opts.FullyQualifiedMessageNames {
		title = string(tt.FullName())
	}
	s := &base.Schema{
		Format:      "enum",
		Title:       title,
		Description: desc,
		Type:        []string{"string"},
		Enum:        children,
		Default:     children[0],
	}
	return string(tt.FullName()), s
}
