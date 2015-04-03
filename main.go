package main

import (
	"flag"
	"fmt"
	pb "github.com/soulplant/runner/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

var serverMode = flag.Bool("server", false, "start the runner server")

type server struct {
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello, %s\n", in.Name),
	}, nil
}

func main() {
	flag.Parse()
	if *serverMode {
		serverMain()
		return
	}
	clientMain()
}

func clientMain() {
	conn, err := grpc.Dial("localhost:1234")
	if err != nil {
		log.Fatalf("dial: %s\n", err)
	}
	client := pb.NewGreeterClient(conn)

	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{"world"})
	if err != nil {
		log.Fatalf("call: %s\n", err)
	}
	log.Printf("Greeting: %s\n", resp.Message)
}

func serverMain() {
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("listen: %s\n", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(l)
}
