# protoc-gen-openapi

- 本项目一个 protobuf openapi 插件的实现
- 本项目早期参考了[protoc-gen-openapi](https://github.com/google/gnostic/tree/master/cmd/protoc-gen-openapi), 由于 gnostic 长时间不更新，并发现了优秀的替代品[protoc-gen-connect-openapi](https://github.com/sudorandom/protoc-gen-connect-openapi),  现在基于[protoc-gen-connect-openapi](https://github.com/sudorandom/protoc-gen-connect-openapi) 做了部分修改
- internal/converter 大部分都是来自于 protoc-gen-connect-openapi, copy.go 部分是自己的实现

## 为什么修改 protoc-gen-connect-openapi
- protoc-gen-connect-openapi 刚好和本项目早期底层使用一致, 为本项目重构提供了很好的思路和想法
- protoc-gen-connect-openapi 是为了 connect-go 生成 openapi 文档, 包含不少 connect 关键词
- protoc-gen-connect-openapi 不支持 grpc service 级别的注解
- protoc-gen-connect-openapi 针对企业内部业务级别的修改无法支持

## 改动部分

- 使用原生的 [google.golang.org/protobuf/compiler/protogen](https://pkg.go.dev/google.golang.org/protobuf/compiler/protogen)
- 添加了 grpc service 级别的注解 [service.proto](./proto/openapiv3/service.proto), [example](./examples/tests/openapiv3annotations/message.proto)
- 添加了 [lava](https://github.com/pubgo/lava) 相关的 header 信息
- 添加了 Error Code 定义 [error.proto](https://github.com/pubgo/funk/blob/master/proto/errorpb/errors.proto)
- 使用 [protobuild](https://github.com/pubgo/protobuild) 进行 protobuf 构建和管理

 ## Installation:

    go install github.com/pubgo/protoc-gen-openapi@latest

## 使用

参考 `make protobuf_test`

## 计划

- 长期兼容 gnostic, 兼容 protoc-gen-connect-openapi, 后期会进行重构
