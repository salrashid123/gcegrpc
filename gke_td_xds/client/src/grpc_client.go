package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"os"

	"github.com/salrashid123/gcegrpc/app/echo"

	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	//	"google.golang.org/grpc/resolver"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"
)

const ()

var (
	conn *grpc.ClientConn
)

func main() {

	address := flag.String("host", "xds:///fe-srv-td:50051", "XDS Servers Listener Name")
	tlsCert := flag.String("tlsCert", "", "tls Certificate")
	serverName := flag.String("servername", "grpc.domain.com", "CACert for server")
	useTLS := flag.Bool("useTLS", true, "do not use TLS Certificate")
	flag.Parse()

	var conn *grpc.ClientConn
	var err error
	if *useTLS {
		var tlsCfg tls.Config
		rootCAs := x509.NewCertPool()
		pem, err := os.ReadFile(*tlsCert)
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
	} else {
		conn, err = grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg "})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("RPC Response: %v %v", i, r)
		time.Sleep(1 * time.Second)
	}

}
