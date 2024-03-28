package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/salrashid123/gcegrpc/app/echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	//"google.golang.org/grpc/codes"
	//healthpb "google.golang.org/grpc/health/grpc_health_v1"
	//"google.golang.org/grpc/status"
)

const ()

var (
	conn *grpc.ClientConn
)

func main() {

	address := flag.String("host", "localhost:50051", "host:port of gRPC server")
	insecure := flag.Bool("insecure", false, "connect without TLS")
	cacert := flag.String("cacert", "CA_crt.pem", "CACert for server")
	serverName := flag.String("servername", "grpc.domain.com", "CACert for server")
	flag.Parse()

	var err error
	if *insecure == true {
		conn, err = grpc.Dial(*address, grpc.WithInsecure())
	} else {

		var tlsCfg tls.Config
		rootCAs := x509.NewCertPool()
		pem, err := ioutil.ReadFile(*cacert)
		if err != nil {
			log.Fatalf("failed to load root CA certificates  error=%v", err)
		}
		if !rootCAs.AppendCertsFromPEM(pem) {
			log.Fatalf("no root CA certs parsed from file ")
		}
		tlsCfg.RootCAs = rootCAs
		tlsCfg.ServerName = *serverName

		ce := credentials.NewTLS(&tlsCfg)
		conn, err = grpc.Dial(*address, grpc.WithTransportCredentials(ce))
	}
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	/*
		ctx, cancel := context.WithTimeout(ctx, 1 * time.Second)
		defer cancel()
		resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "echo.EchoServer"})
		if err != nil {
			log.Fatalf("HealthCheck failed %+v", err)
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			log.Fatalf("service not in serving state: ", resp.GetStatus().String())
		}
		log.Printf("RPC HealthChekStatus:%v", resp.GetStatus())
	*/

	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg "})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		time.Sleep(1 * time.Second)
		log.Printf("RPC Response: %v %v", i, r)
	}

	/*
		stream, err := c.SayHelloStream(ctx, &pb.EchoRequest{Name: "Stream RPC msg"}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("SayHelloStream(_) = _, %v", err)
		}
		for {
			m, err := stream.Recv()
			if err == io.EOF {
				t := stream.Trailer()
				log.Println("Stream Trailer: ", t)
				break
			}
			if err != nil {
				log.Fatalf("SayHelloStream(_) = _, %v", err)
			}
			h, err := stream.Header()
			if err != nil {
				log.Fatalf("stream.Header error _, %v", err)
			}
			log.Printf("Stream Header: ", h)
			log.Printf("Message: ", m.Message)
		}
	*/

}
