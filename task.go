package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type Task struct {
	name string
	cmds []string
}

func parseFile(filename string) ([]Task, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseBytes(data)
}

func parseBytes(data []byte) ([]Task, error) {
	lines := strings.Split(string(data), "\n")
	// for n, line := range lines {
	n := 0
	tasks := []Task{}
	for n < len(lines) {
		line := lines[n]
		n++
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, " ") {
			return nil, fmt.Errorf("line %d: unexpected step, expected task", n)
		}
		name := line
		cmds := []string{}
		line = lines[n]
		n++
		for strings.HasPrefix(line, " ") {
			cmd := strings.TrimLeft(line, " ")
			cmds = append(cmds, cmd)
			if n >= len(lines) {
				break
			}
			line = lines[n]
			n++
		}
		tasks = append(tasks, Task{name, cmds})
	}
	return tasks, nil
}
