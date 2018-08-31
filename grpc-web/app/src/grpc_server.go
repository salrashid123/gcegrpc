package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

var (
	grpcport = flag.String("grpcport", "", "grpcport")
)

const ()

type server struct {
}

func (s *server) SayHelloStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloStreamServer) error {

	log.Println("Got stream:  -->  ")
	ctx := stream.Context()
	//log.Println(ctx)

	var respmdheader = metadata.MD{
		"streamheaderkey": []string{"val"},
	}
	if err := grpc.SendHeader(ctx, respmdheader); err != nil {
		log.Fatalf("grpc.SendHeader(%v, %v) = %v, want %v", ctx, respmdheader, err, nil)
	}

	stream.Send(&echo.EchoReply{Message: "Streaming Response 1 for Request " + in.Name})
	stream.Send(&echo.EchoReply{Message: "Streaming Response 2 for Request " + in.Name})
	stream.Send(&echo.EchoReply{Message: "Streaming Response 3 for Request " + in.Name})

	var respmdfooter = metadata.MD{
		"streamtrailerkey": []string{"val"},
	}
	grpc.SetTrailer(ctx, respmdfooter)

	return nil
}

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {

	log.Println("Got rpc: --> ", in.Name)
	//log.Println(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Println(md["authorization"])
	}

	var respmdheader = metadata.MD{
		"rpcheaderkey": []string{"val"},
	}
	if err := grpc.SendHeader(ctx, respmdheader); err != nil {
		log.Fatalf("grpc.SendHeader(%v, %v) = %v, want %v", ctx, respmdheader, err, nil)
	}
	var respmdfooter = metadata.MD{
		"rpctrailerkey": []string{"val"},
	}
	grpc.SetTrailer(ctx, respmdfooter)

	var h, err = os.Hostname()
	if err != nil {
		log.Fatalf("Unable to get hostname %v", err)
	}
	return &echo.EchoReply{Message: "Hello " + in.Name + "  from hostname " + h}, nil
}

func main() {

	flag.Parse()

	if *grpcport == "" {
		fmt.Fprintln(os.Stderr, "missing -grpcport flag (:50051)")
		flag.Usage()
		os.Exit(2)
	}

	ce, err := credentials.NewServerTLSFromFile("server_crt.pem", "server_key.pem")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	sopts = append(sopts, grpc.Creds(ce))
	s := grpc.NewServer(sopts...)

	echo.RegisterEchoServerServer(s, &server{})
	healthpb.RegisterHealthServer(s, &health.Server{})

	log.Printf("Starting gRPC server on port %v", *grpcport)

	s.Serve(lis)
}
