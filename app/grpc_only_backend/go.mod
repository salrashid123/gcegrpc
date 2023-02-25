module main

go 1.13

require (
	echo v0.0.0
	golang.org/x/net v0.7.0
	google.golang.org/grpc v1.28.0
)

replace echo => ./src/echo
