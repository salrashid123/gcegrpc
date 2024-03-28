module main

go 1.21

require (
	github.com/salrashid123/gcegrpc/app/echo v0.0.0
	golang.org/x/net v0.22.0
	google.golang.org/grpc v1.62.1

)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/salrashid123/gcegrpc/app/echo => ./src/echo
