package main

import (
	//"io/ioutil"
	"log"
	//"os"
	pb "echo"
	"golang.org/x/net/context"
	//"golang.org/x/oauth2"
	//"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	//"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
)

const (
	address = "ip_of_grpc_server:50051"
	//https://developers.google.com/console/help/new/#serviceaccounts
	serviceAccountJSONFile = "YOUR_JSON_CERT_FILE.json"
	userinfo_scope         = "https://www.googleapis.com/auth/userinfo.email"
)

var ()

func main() {

	//https://developers.google.com/identity/protocols/application-default-credentials
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountJSONFile)
	/*
		src, err := google.DefaultTokenSource(oauth2.NoContext, userinfo_scope)
		if err != nil {
			log.Fatalf("Unable to acquire token source: %v", err)
		}
	*/

	// For JWTAccessTokens as authorization header
	/*
		dat, err := ioutil.ReadFile(serviceAccountJSONFile)
		if err != nil {
			log.Fatalf("Unable to read service account file %v", err)
		}
		src, err = google.JWTAccessTokenSourceFromJSON(dat, userinfo_scope)
	*/

	/*
		creds := oauth.TokenSource{src}
		tok, err := creds.Token()
		if err != nil {
			log.Fatalf("Unable to acquire token source: %v", err)
		}
	*/

	//https://github.com/golang/oauth2/issues/127
	//http://stackoverflow.com/questions/26159658/golang-token-validation-error/26287613#26287613
	/*
		if (tok.Extra("id_token") != nil){
			log.Printf("id_token: " , tok.Extra("id_token").(string))
		}
	*/

	ce, err := credentials.NewClientTLSFromFile("CA_crt.pem", "192.168.1.3")
	if err != nil {
		log.Fatalf("Failed to generate credentials %v", err)
	}

	//conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce), grpc.WithPerRPCCredentials(creds))

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(ce))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewEchoServerClient(conn)

	var testMetadata = metadata.MD{
		"sal":  []string{"value1"},
		"key2": []string{"value2"},
	}

	ctx := metadata.NewContext(context.Background(), testMetadata)

	var header, trailer metadata.MD
	r, err := c.SayHello(ctx, &pb.EchoRequest{Name: "unary RPC msg "}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("RPC Response: %v", r)
	log.Printf("Header %v", header)
	log.Printf("Trailer %v", trailer)

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
