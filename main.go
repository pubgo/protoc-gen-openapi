package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/pubgo/protoc-gen-openapi/internal/converter"
	_ "github.com/pubgo/protoc-gen-openapi/internal/logging"
	"github.com/pubgo/protoc-gen-openapi/version"
)

var showVersion = flag.Bool("version", false, "print the version and exit")

func main() {
	flag.Parse()

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
