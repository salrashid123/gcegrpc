package main

import (
	"flag"
	"io"
	"log"
	"net"

	echo "github.com/salrashid123/go-grpc-bazel-docker/echo/echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	grpcport = flag.String("grpcport", ":8080", "grpcport")
)

const ()

type server struct {
}

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (s *server) SayHelloUnary(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {
	log.Println("Got Unary Request: ")
	uid, _ := uuid.NewUUID()
	return &echo.EchoReply{Message: "SayHelloUnary Response " + uid.String()}, nil
}

func (s *server) SayHelloServerStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloServerStreamServer) error {
	log.Println("Got SayHelloServerStream: Request ")
	for i := 0; i < 5; i++ {
		stream.Send(&echo.EchoReply{Message: "SayHelloServerStream Response"})
	}
	return nil
}

func (s server) SayHelloBiDiStream(srv echo.EchoServer_SayHelloBiDiStreamServer) error {
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}
		log.Printf("Got SayHelloBiDiStream %s", req.Name)
		resp := &echo.EchoReply{Message: "SayHelloBiDiStream Server Response"}
		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

func (s server) SayHelloClientStream(stream echo.EchoServer_SayHelloClientStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&echo.EchoReply{Message: "SayHelloClientStream  Response"})
		}
		if err != nil {
			return err
		}
		log.Printf("Got SayHelloClientStream Request: %s", req.Name)
	}
}

func main() {

	flag.Parse()

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(20)}

	s := grpc.NewServer(sopts...)

	echo.RegisterEchoServerServer(s, &server{})
	healthpb.RegisterHealthServer(s, &healthServer{})

	s.Serve(lis)

}
