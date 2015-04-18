package utils

import (
	"log"
	"testing"
)

var str = `
 <name>{{.DomainName}}</name>
  <uuid></uuid>
  <memory>{{.MemorySize}}</memory>
  <vcpu>{{.CPUs}}</vcpu>
`

type data struct {
	DomainName string
	MemorySize uint
	CPUs       uint
}

func TestProcessTemplate(t *testing.T) {
	d := &data{
		DomainName: "test",
		MemorySize: 1000,
		CPUs:       2,
	}
	out, err := ProcessTemplate(str, d)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("DEBUG: %s\n", out)
}
