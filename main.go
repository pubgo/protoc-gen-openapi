package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/pubgo/protoc-gen-openapi/internal/converter"
	_ "github.com/pubgo/protoc-gen-openapi/internal/logging"
	"github.com/pubgo/protoc-gen-openapi/version"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

var showVersion = flag.Bool("version", false, "print the version and exit")

func main() {
	flag.Parse()

	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		var originFiles []*protogen.GeneratedFile

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			originFiles = append(originFiles, gengo.GenerateFile(gen, f))
		}

		for _, f := range originFiles {
			f.Skip()
		}
		return nil
	})

	if *showVersion {
		fmt.Printf("protoc-gen-openapi %s\n", version.FullVersion())
		return
	}

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
