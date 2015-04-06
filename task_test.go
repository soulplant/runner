package main

import (
	"reflect"
	"testing"
)

const input1 = `
build
  ./gen.sh
  go build

test
  go test
`

func TestStuff(t *testing.T) {
	tasks, err := parseBytes([]byte(input1))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := []Task{
		Task{"build", []string{"./gen.sh", "go build"}},
		Task{"test", []string{"go test"}},
	}
	if !reflect.DeepEqual(tasks, expected) {
		t.Errorf("Expected %v, got %v", expected, tasks)
	}
}
