
### gRPC client/server&bidi streaming using Bazel

### Build Image

```
bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
```

### Run Local

```
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- grpcport :8080

bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:8080
```

### Using Docker

```
docker run -p 8080:8080 bazel/greeter_server:greeter_server_image
docker run --net=host bazel/greeter_client:greeter_client_image
```


also see

[Deterministic builds with go + bazel + grpc + docker](https://github.com/salrashid123/go-grpc-bazel-docker)