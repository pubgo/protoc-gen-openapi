vet:
	go vet ./...

protobuf:
	protobuild vendor
	protobuild gen
	rm -rf generator
	mv github.com/pubgo/protoc-gen-openapi/generator generator
	rm -rf github.com

protobuf_test:
	protobuild vendor -c protobuf_test.yaml
	protobuild gen -c protobuf_test.yaml

install:
	go install .
