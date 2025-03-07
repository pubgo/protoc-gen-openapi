// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package generator

import (
	"log"
	"strings"

	v3 "github.com/google/gnostic-models/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/protoc-gen-openapi/generator/wellknown"
)

const (
	protobufValueName = "GoogleProtobufValue"
	protobufAnyName   = "GoogleProtobufAny"
)

type OpenAPIv3Reflector struct {
	conf Configuration

	requiredSchemas []string // Names of schemas which are used through references.
}

// NewOpenAPIv3Reflector creates a new reflector.
func NewOpenAPIv3Reflector(conf Configuration) *OpenAPIv3Reflector {
	return &OpenAPIv3Reflector{
		conf:            conf,
		requiredSchemas: make([]string, 0),
	}
}

func (r *OpenAPIv3Reflector) formatMessageName(message protoreflect.MessageDescriptor) string {
	typeName := fullMessageTypeName(message)
	name := getMessageName(message)
	if !*r.conf.FQSchemaNaming {
		if typeName == ".google.protobuf.Value" {
			name = protobufValueName
		} else if typeName == ".google.protobuf.Any" {
			name = protobufAnyName
		}
	}

	if *r.conf.Naming == "json" {
		if len(name) > 1 {
			name = strings.ToUpper(name[0:1]) + name[1:]
		}

		if len(name) == 1 {
			name = strings.ToLower(name)
		}
	}

	if *r.conf.FQSchemaNaming {
		packageName := string(message.ParentFile().Package())
		name = packageName + "." + name
	}

	return name
}

func (r *OpenAPIv3Reflector) formatFieldName(field protoreflect.FieldDescriptor) string {
	if *r.conf.Naming == "proto" {
		return string(field.Name())
	}

	return field.JSONName()
}

func (r *OpenAPIv3Reflector) responseContentForMessage(message protoreflect.MessageDescriptor) (string, *v3.MediaTypes) {
	typeName := fullMessageTypeName(message)

	if typeName == ".google.protobuf.Empty" {
		return "200", &v3.MediaTypes{}
	}

	if typeName == ".google.api.HttpBody" {
		return "200", wellknown.NewGoogleApiHttpBodyMediaType()
	}

	return "200", wellknown.NewApplicationJsonMediaType(r.schemaOrReferenceForMessage(message))
}

func (r *OpenAPIv3Reflector) schemaReferenceForMessage(message protoreflect.MessageDescriptor) string {
	schemaName := r.formatMessageName(message)
	if !contains(r.requiredSchemas, schemaName) {
		r.requiredSchemas = append(r.requiredSchemas, schemaName)
	}
	return "#/components/schemas/" + schemaName
}

// Returns a full schema for simple types, and a schema reference for complex types that reference
// the definition in `#/components/schemas/`
func (r *OpenAPIv3Reflector) schemaOrReferenceForMessage(message protoreflect.MessageDescriptor) *v3.SchemaOrReference {
	typeName := fullMessageTypeName(message)
	switch typeName {
	case ".google.api.HttpBody":
		return wellknown.NewGoogleApiHttpBodySchema()

	case ".google.protobuf.Timestamp":
		return wellknown.NewGoogleProtobufTimestampSchema()

	case ".google.protobuf.Duration":
		return wellknown.NewGoogleProtobufDurationSchema()

	case ".google.type.Date":
		return wellknown.NewGoogleTypeDateSchema()

	case ".google.type.DateTime":
		return wellknown.NewGoogleTypeDateTimeSchema()

	case ".google.protobuf.FieldMask":
		return wellknown.NewGoogleProtobufFieldMaskSchema()

	case ".google.protobuf.Struct":
		return wellknown.NewGoogleProtobufStructSchema()

	case ".google.protobuf.Empty":
		// Empty is closer to JSON undefined than null, so ignore this field
		return nil //&v3.SchemaOrReference{Oneof: &v3.SchemaOrReference_Schema{Schema: &v3.Schema{Type: "null"}}}

	case ".google.protobuf.BoolValue":
		return wellknown.NewBooleanSchema()

	case ".google.protobuf.BytesValue":
		return wellknown.NewBytesSchema()

	case ".google.protobuf.Int32Value", ".google.protobuf.UInt32Value":
		return wellknown.NewIntegerSchema(getValueKind(message))

	case ".google.protobuf.StringValue", ".google.protobuf.Int64Value", ".google.protobuf.UInt64Value":
		return wellknown.NewStringSchema()

	case ".google.protobuf.FloatValue", ".google.protobuf.DoubleValue":
		return wellknown.NewNumberSchema(getValueKind(message))

	default:
		ref := r.schemaReferenceForMessage(message)
		return &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Reference{Reference: &v3.Reference{XRef: ref}},
		}
	}
}

func (r *OpenAPIv3Reflector) schemaOrReferenceForField(field *protogen.Field, desc protoreflect.FieldDescriptor) *v3.SchemaOrReference {
	if desc == nil {
		desc = field.Desc
	}

	var kindSchema *v3.SchemaOrReference
	kind := desc.Kind()
	switch kind {
	case protoreflect.MessageKind:
		if desc.IsMap() {
			// This means the field is a map, for example:
			//   map<string, value_type> map_field = 1;
			//
			// The map ends up getting converted into something like this:
			//   message MapFieldEntry {
			//     string key = 1;
			//     value_type value = 2;
			//   }
			//
			//   repeated MapFieldEntry map_field = N;
			//
			// So we need to find the `value` field in the `MapFieldEntry` message and
			// then return a MapFieldEntry schema using the schema for the `value` field
			return wellknown.NewGoogleProtobufMapFieldEntrySchema(r.schemaOrReferenceForField(field, desc.MapValue()))
		} else {
			kindSchema = r.schemaOrReferenceForMessage(desc.Message())
		}

	case protoreflect.StringKind:
		kindSchema = wellknown.NewStringSchema()

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		kindSchema = wellknown.NewIntegerSchema(kind.String())

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
		protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		kindSchema = wellknown.NewStringSchema()

	case protoreflect.EnumKind:
		kindSchema = wellknown.NewEnumSchema(r.conf.EnumType, field)

	case protoreflect.BoolKind:
		kindSchema = wellknown.NewBooleanSchema()

	case protoreflect.FloatKind, protoreflect.DoubleKind:
		kindSchema = wellknown.NewNumberSchema(kind.String())

	case protoreflect.BytesKind:
		kindSchema = wellknown.NewBytesSchema()

	default:
		log.Printf("(TODO) Unsupported field type: %+v", fullMessageTypeName(field.Desc.Message()))
	}

	if field.Desc.IsList() {
		kindSchema = wellknown.NewListSchema(kindSchema)
	}

	return kindSchema
}
