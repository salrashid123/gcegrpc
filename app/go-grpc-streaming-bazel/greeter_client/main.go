package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	echo "github.com/salrashid123/go-grpc-bazel-docker/echo/echo"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const ()

var (
	conn *grpc.ClientConn
)

func main() {

	address := flag.String("host", "localhost:8080", "host:port of gRPC server")
	cacert := flag.String("cacert", "", "caCertificateFile")
	flag.Parse()

	var err error

	if *cacert == "" {
		conn, err = grpc.Dial(*address, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(*cacert, "")
		if err != nil {
			log.Fatalf("failed  to get cacerts %+v", err)
		}
		conn, err = grpc.Dial(*address, grpc.WithTransportCredentials(creds))
	}

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)
	ctx := context.Background()

	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: "echo.EchoServer"})
	if err != nil {
		log.Fatalf("HealthCheck failed %+v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		log.Fatalf("service not in serving state: ", resp.GetStatus().String())
	}
	log.Printf("RPC HealthChekStatus:%v", resp.GetStatus())

	/// UNARY
	for i := 0; i < 5; i++ {
		r, err := c.SayHelloUnary(ctx, &echo.EchoRequest{Name: fmt.Sprintf("Unary Request %d", i)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		time.Sleep(1 * time.Second)
		log.Printf("Unary Response: %v [%v]", i, r)
	}

	// ********CLIENT

	cstream, err := c.SayHelloClientStream(context.Background())

	if err != nil {
		log.Fatalf("%v.SayHelloClientStream(_) = _, %v", c, err)
	}

	for i := 1; i < 5; i++ {
		if err := cstream.Send(&echo.EchoRequest{Name: fmt.Sprintf("client stream RPC %d ", i)}); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%v.Send(%v) = %v", cstream, i, err)
		}
	}

	creply, err := cstream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", cstream, err, nil)
	}
	log.Printf(" Got SayHelloClientStream  [%s]", creply.Message)

	/// ***** SERVER
	stream, err := c.SayHelloServerStream(ctx, &echo.EchoRequest{Name: "Stream RPC msg"})
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

		log.Printf("Message: [%s]", m.Message)
	}

	/// ********** BIDI

	done := make(chan bool)
	stream, err = c.SayHelloBiDiStream(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}
	ctx = stream.Context()

	go func() {
		for i := 1; i <= 10; i++ {
			req := echo.EchoRequest{Name: "Bidirectional CLient RPC msg "}
			if err := stream.SendMsg(&req); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("Response: [%s] ", resp.Message)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done

}
