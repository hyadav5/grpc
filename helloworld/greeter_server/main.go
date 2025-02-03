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

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

var streamResp = []string{"I am good", "what about you ?"}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("----------------- Unary RPC ------------------------------\n")
	fmt.Printf("Received unary request %v\n", request)
	resp := pb.HelloReply{Message: "I am good"}
	fmt.Printf("Sending unary response %v\n", &resp)
	return &resp, nil
}

func (s *server) LotsOfReplies(request *pb.HelloRequest, stream pb.Greeter_LotsOfRepliesServer) error {
	fmt.Printf("--------------- Server streaming RPC --------------------------------\n")
	fmt.Printf("Received unary request %v\n", request)
	resp := pb.HelloReply{Message: ""}
	for _, msg := range streamResp {
		resp.Message = msg
		if err := stream.Send(&resp); err != nil {
			fmt.Printf("Failed to send response %v\n", err)
			return err
		}
		fmt.Printf("Sending streaming response %v\n", &resp)
	}
	return nil
}

func (s *server) LotsOfGreetings(stream pb.Greeter_LotsOfGreetingsServer) error {
	fmt.Printf("------------------- Client streaming RPC ----------------------------\n")
	for {
		request, err := stream.Recv()
		if err == io.EOF { // if it iss io.EOF the message stream has ended.
			fmt.Printf("Received all stream requests.\n")
			resp := pb.HelloReply{Message: "Too much greetings"}
			fmt.Printf("Sending unary response %v\n", &resp)
			_ = stream.SendAndClose(&resp)
			break
		}
		if err != nil {
			fmt.Printf("Failed to received stream request %v\n", err)
			return err
		}
		if request != nil { // If the request is nil, the stream is still good and it can continue reading.
			fmt.Printf("Received streaming request %v\n", request)
		}
	}
	return nil
}

func (s *server) PingPong(streamBiDi pb.Greeter_PingPongServer) error {
	fmt.Printf("--------------- Bidirectional streaming RPC --------------------------------\n")
	for {
		request, err := streamBiDi.Recv()
		if err == io.EOF { // if it iss io.EOF the message stream has ended.
			fmt.Printf("Received all stream requests.\n")
			break
		}
		if err != nil {
			fmt.Printf("Failed to received stream request %v\n", err)
			return err
		}
		if request != nil { // If the request is nil, the stream is still good and it can continue reading.
			fmt.Printf("Received streaming request %v\n", request)
		}
		if request.Message == "bye" {
			arrRequest := []string{"I am good.", "Life is also good."}
			fmt.Printf("Received all stream requests.\n")
			for _, msg := range arrRequest {
				resp := pb.HelloReply{Message: msg}
				fmt.Printf("Sending streaming response %v\n", &resp)
				err = streamBiDi.Send(&resp)
				if err != nil {
					fmt.Printf("Failed to send streaming response %v\n", err)
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	sopts := []grpc.ServerOption{}
	sopts = append(sopts,
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    time.Duration(10) * time.Second,
				Timeout: time.Duration(10) * time.Second,
			},
		),
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             time.Duration(5) * time.Second,
				PermitWithoutStream: true,
			},
		),
	)

	s := grpc.NewServer(sopts...)
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("Server Started at %v, listening at port %v", lis.Addr(), time.Now())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
