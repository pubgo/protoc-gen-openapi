package converter

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/pb33f/libopenapi"
	base "github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors/errcheck"
	"github.com/pubgo/funk/recovery"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/protoc-gen-openapi/internal/converter/gnostic"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/util"
)

func Convert(gen *protogen.Plugin, cfg options.Config) error {
	defer recovery.Exit()

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

func mergeTags(tags []*base.Tag) []*base.Tag {
	if len(tags) == 0 {
		return tags
	}

	tagList := make([]*base.Tag, 0, len(tags))
	found := make(map[string]*base.Tag)

	for _, tag := range tags {
		if found[tag.Name] == nil {
			found[tag.Name] = tag
			tagList = append(tagList, tag)
			continue
		}

		if tag.Description != "" {
			found[tag.Name].Description = tag.Description
		}

		if tag.ExternalDocs != nil {
			found[tag.Name].ExternalDocs = tag.ExternalDocs
		}

		if tag.Extensions != nil {
			found[tag.Name].Extensions = tag.Extensions
		}
	}

	return tagList
}

func specToFile(opts options.Options, spec *v3.Document) (string, error) {
	switch opts.Format {
	case "yaml":
		return string(spec.RenderWithIndention(2)), nil
	case "json":
		b := assert.Must1(spec.RenderJSON("  "))
		return string(b), nil
	default:
		return "", fmt.Errorf("unknown format: %s", opts.Format)
	}
}

func appendToSpec(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) (gErr error) {
	gnostic.SpecWithFileAnnotations(spec, fd)
	components, err := fileToComponents(opts, fd)
	if errcheck.Check(&gErr, err) {
		return
	}

	initializeDoc(spec)
	initializeComponents(components)
	appendServiceDocs(opts, spec, fd)
	util.AppendComponents(spec, components)

	if errcheck.Check(&gErr, addPathItemsFromFile(opts, fd, spec.Paths)) {
		return
	}
	spec.Tags = append(spec.Tags, fileToTags(opts, fd)...)
	return nil
}

func appendServiceDocs(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) {
	if !opts.WithServiceDescriptions {
		return
	}

	var builder strings.Builder
	if spec.Info.Description != "" {
		builder.WriteString(spec.Info.Description)
		builder.WriteString("\n\n")
	}

	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}

		builder.WriteString("## ")
		builder.WriteString(string(service.FullName()))
		builder.WriteString("\n\n")

		loc := fd.SourceLocations().ByDescriptor(service)
		serviceComments := util.FormatComments(loc)
		if serviceComments != "" {
			builder.WriteString(serviceComments)
			builder.WriteString("\n\n")
		}
	}

	spec.Info.Description = strings.TrimSpace(builder.String())
}

func initializeDoc(doc *v3.Document) {
	slog.Debug("initializeDoc")
	if doc.Version == "" {
		doc.Version = "3.1.0"
	}
	if doc.Paths == nil {
		doc.Paths = &v3.Paths{}
	}
	if doc.Paths.PathItems == nil {
		doc.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Paths.Extensions == nil {
		doc.Paths.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Info == nil {
		doc.Info = &base.Info{}
	}
	if doc.Paths == nil {
		doc.Paths = &v3.Paths{}
	}
	if doc.Paths.PathItems == nil {
		doc.Paths.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Paths.Extensions == nil {
		doc.Paths.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Security == nil {
		doc.Security = []*base.SecurityRequirement{}
	}
	if doc.Extensions == nil {
		doc.Extensions = orderedmap.New[string, *yaml.Node]()
	}
	if doc.Webhooks == nil {
		doc.Webhooks = orderedmap.New[string, *v3.PathItem]()
	}
	if doc.Index == nil {
		doc.Index = &index.SpecIndex{}
	}
	if doc.Rolodex == nil {
		doc.Rolodex = &index.Rolodex{}
	}
	if doc.Components == nil {
		doc.Components = &v3.Components{}
	}
	initializeComponents(doc.Components)
}

func initializeComponents(components *v3.Components) {
	if components.Schemas == nil {
		components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
	}
	if components.Responses == nil {
		components.Responses = orderedmap.New[string, *v3.Response]()
	}
	if components.Parameters == nil {
		components.Parameters = orderedmap.New[string, *v3.Parameter]()
	}
	if components.Examples == nil {
		components.Examples = orderedmap.New[string, *base.Example]()
	}
	if components.RequestBodies == nil {
		components.RequestBodies = orderedmap.New[string, *v3.RequestBody]()
	}
	if components.Headers == nil {
		components.Headers = orderedmap.New[string, *v3.Header]()
	}
	if components.SecuritySchemes == nil {
		components.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}
	if components.Links == nil {
		components.Links = orderedmap.New[string, *v3.Link]()
	}
	if components.Callbacks == nil {
		components.Callbacks = orderedmap.New[string, *v3.Callback]()
	}
	if components.Extensions == nil {
		components.Extensions = orderedmap.New[string, *yaml.Node]()
	}
}
