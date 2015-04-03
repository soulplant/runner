package main

import (
	"flag"
	"fmt"
	pb "github.com/soulplant/runner/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
)

var serverFlag = flag.Bool("server", false, "start the runner server")
var commandFlag = flag.String("cmd", "", "command to execute")

const prompt = "$ "

type server struct {
	l       sync.Mutex
	running *exec.Cmd
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello, %s\n", in.Name),
	}, nil
}

func (s *server) Run(req *pb.RunRequest, resp pb.Greeter_RunServer) error {
	f, err := ioutil.TempFile("", "runner-")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", req.Command)
	cmd := exec.Command("bash", "-c", req.Command)
	writer := io.MultiWriter(f, os.Stdout)
	cmd.Stdout = writer
	cmd.Stderr = writer
	err = s.start(cmd)
	if err != nil {
		return err
	}
	resp.Send(&pb.RunReply{Filename: f.Name()})
	fmt.Printf(prompt)
	return nil
}

func (s *server) start(cmd *exec.Cmd) error {
	s.l.Lock()
	defer s.l.Unlock()
	if s.running != nil {
		s.running.Process.Kill()
		s.running = nil
	}
	err := cmd.Start()
	if err != nil {
		return err
	}
	s.running = cmd
	return nil
}

func writeSync(f *os.File, text string) {
	var written int
	for written < len(text) {
		n, err := f.Write([]byte(text[written:]))
		if err != nil {
			return
		}
		log.Printf("wrote %d byte(s)\n", n)
		written += n
	}
}

func main() {
	flag.Parse()
	if *serverFlag {
		serverMain()
	} else {
		clientMain()
	}
}

func clientMain() {
	conn, err := grpc.Dial("localhost:1234")
	if err != nil {
		log.Fatalf("dial: %s\n", err)
	}
	client := pb.NewGreeterClient(conn)
	if commandFlag == nil {
		log.Fatalf("cmd argument required\n")
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
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("listen: %s\n", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	fmt.Printf(prompt)
	s.Serve(l)
}
