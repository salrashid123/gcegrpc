package main

import (
	"echo"
	"flag"

	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	//	"google.golang.org/grpc/resolver"
	_ "google.golang.org/grpc/xds"
)

const ()

var (
	conn *grpc.ClientConn
)

func main() {

	address := flag.String("host", "xds:///fe-srv-td:50051", "XDS Servers Listener Name")
	flag.Parse()
	conn, err := grpc.Dial(*address, grpc.WithInsecure())

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
