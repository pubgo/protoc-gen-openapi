package googleapi

import (
	"github.com/pubgo/protoc-gen-openapi/generator"
	"github.com/pubgo/protoc-gen-openapi/internal/converter/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GetSrvOptions(opts options.Options, serviceDescriptor protoreflect.ServiceDescriptor) *generator.Service {
	srvOpts := serviceDescriptor.Options()
	if srvOpts == nil || !proto.HasExtension(srvOpts, generator.E_Service) {
		return nil
	}

	srv, ok := proto.GetExtension(srvOpts, generator.E_Service).(*generator.Service)
	if !ok {
		return nil
	}

	return srv
}
