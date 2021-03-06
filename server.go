package main

import (
	"fmt"
	pb "github.com/soulplant/runner/proto"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	prompt = "$ "
)

type server struct {
	l       sync.Mutex
	running *exec.Cmd
}

func (s *server) printPrompt() {
	fmt.Printf("%s", prompt)
}

func (s *server) Run(ctx context.Context, req *pb.RunRequest) (*pb.RunReply, error) {
	if s.cancel() {
		fmt.Printf("<interrupted>\n")
		s.printPrompt()
	}
	f, err := ioutil.TempFile("", "runner-")
	if err != nil {
		return nil, err
	}
	commandString := strings.Join(req.Command, " && ")
	fmt.Printf("%s\n", commandString)
	cmd := exec.Command("bash", "-c", commandString)
	writer := io.MultiWriter(f, os.Stdout)
	cmd.Stdout = writer
	cmd.Stderr = writer
	err = s.start(cmd)
	if err != nil {
		return nil, err
	}
	go func() {
		cmd.Process.Wait()
		if s.cancel() {
			s.printPrompt()
		}
	}()
	// TODO(koz): Don't leak this file, rather associate it with the running
	// process and close it when the process terminates.
	return &pb.RunReply{Filename: f.Name()}, nil
}

func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListReply, error) {
	commands := s.getRunningCommands()
	return &pb.ListReply{
		Command: commands,
	}, nil
}

func (s *server) getRunningCommands() []string {
	s.l.Lock()
	defer s.l.Unlock()
	result := []string{}
	if s.running != nil {
		cmd := strings.Join(s.running.Args[2:], " ")
		result = append(result, cmd)
	}
	return result
}

func (s *server) cancel() bool {
	s.l.Lock()
	defer s.l.Unlock()
	if s.running != nil {
		s.running.Process.Kill()
		s.running = nil
		return true
	}
	return false
}

func (s *server) start(cmd *exec.Cmd) error {
	s.l.Lock()
	defer s.l.Unlock()
	err := cmd.Start()
	if err != nil {
		return err
	}
	s.running = cmd
	return nil
}
