package main

import (
	"flag"
	"fmt"
	"log"

	"net/http"
	"os"
	"echo"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	grpcport = flag.String("grpcport", "", "grpcport")
	insecure = flag.Bool("insecure", false, "startup without TLS")

	hs *health.Server

	conn *grpc.ClientConn
)

const (
	address string = ":50051"
)

type server struct {
}

func isGrpcRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc")
}

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {

	log.Println("Got rpc: --> ", in.Name)

	var h, err = os.Hostname()
	if err != nil {
		log.Fatalf("Unable to get hostname %v", err)
	}
	return &echo.EchoReply{Message: "Hello " + in.Name + "  from hostname " + h}, nil
}

func (s *server) SayHelloStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloStreamServer) error {

	log.Println("Got stream:  -->  ")
	ctx := stream.Context()
	log.Println(ctx)

	var respmdheader = metadata.MD{
		"streamheaderkey": []string{"val"},
	}
	if err := grpc.SendHeader(ctx, respmdheader); err != nil {
		log.Fatalf("grpc.SendHeader(%v, %v) = %v, want %v", ctx, respmdheader, err, nil)
	}

	stream.Send(&echo.EchoReply{Message: "Msg1 " + in.Name})
	stream.Send(&echo.EchoReply{Message: "Msg2 " + in.Name})

	var respmdfooter = metadata.MD{
		"streamtrailerkey": []string{"val"},
	}
	grpc.SetTrailer(ctx, respmdfooter)

	return nil
}


func hchandler(w http.ResponseWriter, r *http.Request) {

	//log.Print("service not in serving state: ")
	//hs.SetServingStatus("echo.EchoServer", healthpb.HealthCheckResponse_NOT_SERVING)
	//http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)

	fmt.Fprint(w, "ok")
}

type healthServer struct{}
// Check is used for gRPC health checks
func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
// Watch is not implemented
func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func main() {

	flag.Parse()

	if *grpcport == "" {
		fmt.Fprintln(os.Stderr, "missing -grpcport flag (:50051)")
		flag.Usage()
		os.Exit(2)
	}

	ce, err := credentials.NewClientTLSFromFile("server_crt.pem", "")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	if *insecure == true {
		conn, err = grpc.Dial(address, grpc.WithInsecure())
	} else {
		conn, err = grpc.Dial(address, grpc.WithTransportCredentials(ce))
	}
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	http.HandleFunc("/", hchandler)
	http.HandleFunc("/_ah/health", hchandler)

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	s := grpc.NewServer(sopts...)

	echo.RegisterEchoServerServer(s, &server{})

	healthpb.RegisterHealthServer(s, &healthServer{})

	muxHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGrpcRequest(r) {
			s.ServeHTTP(w, r)
			return
		}
		http.DefaultServeMux.ServeHTTP(w, r)
	})
	log.Printf("Starting gRPC server on port %v", *grpcport)


	if *insecure == false {
		log.Fatal(http.ListenAndServeTLS(*grpcport, "server_crt.pem", "server_key.pem", h2c.NewHandler(muxHandler, &http2.Server{})))
	}
	log.Fatal(http.ListenAndServe(*grpcport, h2c.NewHandler(muxHandler, &http2.Server{})))
}
