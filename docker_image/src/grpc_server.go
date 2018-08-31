package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var ()

const (
	grpcport = ":50051"
	httpport = ":8080"
	//logFile = "/var/log/app_engine/custom_logs/grpcserver.log"
	logFile = "/tmp/grpcserver.log"
)

type server struct {
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

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {

	log.Println("Got rpc: --> ", in.Name)
	log.Println(ctx)
	md, ok := metadata.FromContext(ctx)
	if ok {
		// authorization header from context for oauth2 at client.
		// Verify as access_token to oauth2/tokeninfo
		// https://developers.google.com/identity/protocols/OAuth2UserAgent#tokeninfo-validation
		// https://developers.google.com/identity/protocols/OAuth2ServiceAccount
		// -----------------------------------------------------------------------------
		// or if the id_token is sent in, verify digital signature
		// https://developers.google.com/identity/protocols/OpenIDConnect?hl=en#validatinganidtoken
		// https://github.com/golang/oauth2/issues/127
		// http://stackoverflow.com/questions/26159658/golang-token-validation-error/26287613#26287613
		log.Println(md["authorization"])
		//log.Println(md["sal"])
	}

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

	return &echo.EchoReply{Message: "Hello " + in.Name}, nil
}

func fronthandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Main Handler")
	fmt.Fprint(w, "hello world")
}

func healthhandler(w http.ResponseWriter, r *http.Request) {
	log.Println("heathcheck...")
	fmt.Fprint(w, "ok")
}

func main() {

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	http.HandleFunc("/", fronthandler)
	http.HandleFunc("/_ah/health", healthhandler)
	go http.ListenAndServe(httpport, nil)

	ce, err := credentials.NewServerTLSFromFile("server_crt.pem", "server_key.pem")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	lis, err := net.Listen("tcp", grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	sopts = append(sopts, grpc.Creds(ce))
	s := grpc.NewServer(sopts...)

	echo.RegisterEchoServerServer(s, &server{})
	log.Printf("Starting gRPC server on port %v", grpcport)

	s.Serve(lis)
}
