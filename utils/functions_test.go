package utils

import (
	"log"
	"testing"
)

type data struct {
	DomainName string
	MemorySize uint
	Cpus       int
}

func TestProcessTemplate(t *testing.T) {
	d := &data{
		DomainName: "test",
		MemorySize: 1000,
		Cpus:       2,
	}
	out, err := ProcessTemplate(str, d)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("DEBUG: %s\n", out)
}

var str = `
 <name>{{ .DomainName }}</name>
  <uuid></uuid>
  <memory>{{ .MemorySize }}</memory>
  <currentMemory>{{ .MemorySize }}</currentMemory>
  <vcpu>{{ .Cpus }}</vcpu>
`
