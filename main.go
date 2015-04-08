package main

import (
	"errors"
	"flag"
	"fmt"
	pb "github.com/soulplant/runner/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

type client struct {
	conn   *grpc.ClientConn
	client pb.GreeterClient
}

func NewClient() (*client, error) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	c := pb.NewGreeterClient(conn)
	return &client{conn, c}, nil
}

func (c *client) List() ([]string, error) {
	reply, err := c.client.List(context.Background(), &pb.ListRequest{})
	if err != nil {
		return nil, err
	}
	return reply.Command, nil
}

func (c *client) Run(t Task) (string, error) {
	req := pb.RunRequest{}
	req.Command = t.Cmds
	reply, err := c.client.Run(context.Background(), &req)
	if err != nil {
		return "", err
	}
	if reply.Error != "" {
		return "", errors.New(reply.Error)
	}
	return reply.Filename, nil
}

func clientMain() {
	tasks, err := ParseFile("tasks")
	if err != nil {
		panic(err)
	}

	client, err := NewClient()
	if err != nil {
		log.Fatalf("NewClient: %s\n", err)
	}
	if *commandFlag == "" {
		cmds, err := client.List()
		if err != nil {
			log.Fatalf("couldn't list jobs: %s\n", err)
		}
		fmt.Printf("%d jobs running\n", len(cmds))
		for _, cmd := range cmds {
			fmt.Printf("> %s\n", cmd)
		}
		return
	}
	toRun := tasks[0]
	for _, task := range tasks {
		if task.Name == *commandFlag {
			toRun = task
		}
	}
	filename, err := client.Run(toRun)
	if err != nil {
		log.Fatalf("Run: %s\n", err)
	}
	fmt.Printf("command output in %s\n", filename)
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
