package main

import (
	"flag"
	"fmt"
	pb "github.com/soulplant/runner/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

var serverFlag = flag.Bool("server", false, "start the runner server")
var commandFlag = flag.String("cmd", "", "command to execute")

const (
	port = 1234
)

func main() {
	flag.Parse()
	if *serverFlag {
		serverMain()
	} else {
		clientMain()
	}
}

func clientMain() {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("dial: %s\n", err)
	}
	client := pb.NewGreeterClient(conn)
	if *commandFlag == "" {
		reply, err := client.List(context.Background(), &pb.ListRequest{})
		if err != nil {
			log.Fatalf("couldn't list jobs: %s\n", err)
		}
		fmt.Printf("%d jobs running\n", len(reply.Command))
		for _, cmd := range reply.Command {
			fmt.Printf("> %s\n", cmd)
		}
		return
	}
	c, err := client.Run(context.Background(), &pb.RunRequest{*commandFlag})
	if err != nil {
		log.Fatalf("call: %s\n", err)
	}
	for {
		reply, err := c.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return
		}
		if reply.Error != "" {
			fmt.Printf("error: %s\n", err)
			return
		}
		fmt.Printf("command output in %s\n", reply.Filename)
	}
}

func serverMain() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("listen: %s\n", err)
	}
	s := grpc.NewServer()
	serv := &server{}
	serv.printPrompt()
	pb.RegisterGreeterServer(s, serv)
	s.Serve(l)
}
