FROM golang:1.4

RUN apt-get update -y && apt-get install -y build-essential wget unzip autoconf autogen libtool

RUN ldconfig

WORKDIR /tmp/
RUN cd /tmp/ && git clone https://github.com/google/protobuf.git && cd protobuf && ./autogen.sh && ./configure && make && make install
RUN ldconfig

WORKDIR /go/

RUN go get github.com/golang/protobuf/proto
RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc
RUN go get github.com/bradfitz/http2
RUN go get golang.org/x/oauth2
RUN go get google.golang.org/cloud/compute/metadata

ADD . /go/
RUN protoc --go_out=plugins=grpc:. src/echo/echo.proto

RUN export GOBIN=/go/bin && go install src/grpc_server.go
EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["/go/bin/grpc_server"]
