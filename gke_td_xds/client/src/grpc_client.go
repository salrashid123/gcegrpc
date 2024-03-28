package main

import (
	"flag"

	"github.com/salrashid123/gcegrpc/app/echo"

	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	//	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/xds"
)

const ()

var (
	conn *grpc.ClientConn
)

func main() {

	address := flag.String("host", "xds:///fe-srv-td:50051", "XDS Servers Listener Name")
	flag.Parse()
	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	for i := 0; i < 15; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg "})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("RPC Response: %v %v", i, r)
		time.Sleep(1 * time.Second)
	}

}
