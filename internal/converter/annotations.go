package converter

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/protoc-gen-openapi/internal/converter/gnostic"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/googleapi"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/protovalidate"
)

type annotator struct{}

func (*annotator) AnnotateMessage(opts options.Options, schema *base.Schema, desc protoreflect.MessageDescriptor) *base.Schema {
	schema = protovalidate.SchemaWithMessageAnnotations(opts, schema, desc)
	schema = gnostic.SchemaWithSchemaAnnotations(schema, desc)
	return schema
}

func (*annotator) AnnotateField(opts options.Options, schema *base.Schema, desc protoreflect.FieldDescriptor, onlyScalar bool) *base.Schema {
	schema = protovalidate.SchemaWithFieldAnnotations(opts, schema, desc, onlyScalar)
	schema = gnostic.SchemaWithPropertyAnnotations(schema, desc)
	schema = googleapi.SchemaWithPropertyAnnotations(opts, schema, desc)
	return schema
}

func (*annotator) AnnotateFieldReference(opts options.Options, parent *base.Schema, desc protoreflect.FieldDescriptor) *base.Schema {
	parent = protovalidate.PopulateParentProperties(opts, parent, desc)
	return parent
}
