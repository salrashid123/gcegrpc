package main

import (
	"echo"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

var (
	httpport = flag.String("httpport", "", "httpport")
)

const ()

type server struct {
}

func backendLBhandler(w http.ResponseWriter, r *http.Request) {

	log.Println("BackendLB Handler")

	ctx := context.Background()

	ce, err := credentials.NewClientTLSFromFile("server_crt.pem", "")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	conn, err := grpc.Dial("dns:///be-srv-lb.default.svc.cluster.local:50051", grpc.WithTransportCredentials(ce), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer conn.Close()

	c := echo.NewEchoServerClient(conn)

	backendList := []string{}
	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg "})
		if err != nil {
			http.Error(w, "Coudl not greet ", http.StatusInternalServerError)
		}
		time.Sleep(1 * time.Second)
		if (r != nil) {
			backendList = append(backendList, r.Message)
		}
	}

	var h, err2 = os.Hostname()
	if err2 != nil {
		http.Error(w, "Unable to get hostname %v", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")

	str, err := json.Marshal(backendList)
	if err != nil {
		fmt.Println("Error encoding JSON")
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

	fmt.Fprint(w, "{ \"frontend\": \""+h+"\", \"backends\": "+string(str)+"}")
}

func backendhandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Backend Handler")

	ctx := context.Background()

	ce, err := credentials.NewClientTLSFromFile("server_crt.pem", "")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	address := fmt.Sprintf("be-srv.default.svc.cluster.local:50051")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce))

	defer conn.Close()

	c := echo.NewEchoServerClient(conn)

	backendList := []string{}
	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg "})
		if err != nil {
			http.Error(w, "Coudl not greet ", http.StatusInternalServerError)
		}
		time.Sleep(1 * time.Second)
		backendList = append(backendList, r.Message)
	}

	var h, err2 = os.Hostname()
	if err2 != nil {
		http.Error(w, "Unable to get hostname %v", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	str, err := json.Marshal(backendList)
	if err != nil {
		fmt.Println("Error encoding JSON")
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}

	fmt.Fprint(w, "{ \"frontend\": \""+h+"\", \"backends\": "+string(str)+"}")
}

func fronthandler(w http.ResponseWriter, r *http.Request) {
	log.Println("/ called")
	fmt.Fprint(w, "ok")
}

func healthhandler(w http.ResponseWriter, r *http.Request) {
	log.Println("heathcheck...")
	fmt.Fprint(w, "ok")
}

func main() {

	flag.Parse()

	if *httpport == "" {
		fmt.Fprintln(os.Stderr, "missing -httpport flag (:8081)")
		flag.Usage()
		os.Exit(2)
	}

	http.HandleFunc("/", fronthandler)
	http.HandleFunc("/backend", backendhandler)
	http.HandleFunc("/backendlb", backendLBhandler)
	http.HandleFunc("/_ah/health", healthhandler)

	srv := &http.Server{
		Addr: *httpport,
	}
	http2.ConfigureServer(srv, &http2.Server{})
	err := srv.ListenAndServeTLS("server_crt.pem", "server_key.pem")
	//go srv.ListenAndServe()
	//go http.ListenAndServe(*httpport, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
