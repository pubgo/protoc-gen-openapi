vet:
	go vet ./...

pb:
	protobuild vendor

pb_test:
	protobuild -c protobuf_test.yaml vendor
	protobuild -c protobuf_test.yaml gen

install_gnostic:
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.0

protobuf:
	protobuild vendor
	protobuild gen
	rm -rf generator
	mv github.com/pubgo/protoc-gen-openapi/generator generator
	rm -rf github.com

protobuf-test:
	protobuild vendor -c protobuf_test.yaml
	protobuild gen -c protobuf_test.yaml

ALL: generate test install
PHONY: test install buf-generate

PROTO_FILES=$(shell find internal/converter/testdata -type f -name '*.proto')

generate: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	go generate ./...
	go generate ./internal/converter/testdata

test: generate
	go test -coverprofile=coverage.out -coverpkg=./internal/...,./converter/... ./...
	# To see coverage report:
	# go tool cover -html=coverage.out

install:
	go install

buf-generate: install
	buf generate --path internal/
