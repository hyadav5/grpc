/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             10 * time.Second, // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	greeterClient := pb.NewGreeterClient(conn)

	fmt.Printf("----------------- Unary RPC ------------------------------\n")
	req := pb.HelloRequest{Message: "hey hiii, how are you ?"}
	fmt.Printf("Send unary request %v\n", &req)
	resp, err := greeterClient.SayHello(context.Background(), &req, []grpc.CallOption{}...)
	fmt.Printf("Receives unary response %v\n", resp)

	fmt.Printf("--------------- Server streaming RPC --------------------------------\n")
	fmt.Printf("Sending unary request %v\n", &req)
	streamResp, err := greeterClient.LotsOfReplies(context.Background(), &req, []grpc.CallOption{}...)
	for {
		resp, err = streamResp.Recv()
		if err != nil && err != io.EOF {
			fmt.Printf("Failed to receive streaming response %v\n", err)
			return
		}
		if err == io.EOF {
			fmt.Printf("Received all streaming responses. Bye.\n")
			break
		}
		if resp != nil {
			fmt.Printf("Received streaming response %v\n", resp)
		}
	}

	fmt.Printf("------------------- Client streaming RPC ----------------------------\n")
	streamRequest, _ := greeterClient.LotsOfGreetings(context.Background(), []grpc.CallOption{}...)
	arrRequest := []string{"hey hiii, how are you ?", "how is life?"}
	for _, msg := range arrRequest {
		req := pb.HelloRequest{Message: msg}
		fmt.Printf("Sending streaming request %v\n", &req)
		err = streamRequest.Send(&req)
		if err != nil {
			fmt.Printf("Failed to send streaming request %v\n", err)
			return
		}
	}
	resp, err = streamRequest.CloseAndRecv()
	fmt.Printf("Received unary response %v\n", resp)

	fmt.Printf("--------------- Bidirectional streaming RPC --------------------------------\n")
	arrRequest = append(arrRequest, "bye")
	streamBiDi, _ := greeterClient.PingPong(context.Background(), []grpc.CallOption{}...)

	for _, msg := range arrRequest {
		req := pb.HelloRequest{Message: msg}
		fmt.Printf("Sending streaming request %v\n", &req)
		err = streamBiDi.Send(&req)
		if err != nil {
			fmt.Printf("Failed to send streaming request %v\n", err)
			return
		}
	}
	for {
		resp, err = streamBiDi.Recv()
		if err != nil && err != io.EOF {
			fmt.Printf("Failed to receive streaming response %v\n", err)
			return
		}
		if err == io.EOF {
			fmt.Printf("Received all streaming responses. Bye.\n")
			break
		}
		if resp != nil {
			fmt.Printf("Received streaming response %v\n", resp)
		}
	}
}
