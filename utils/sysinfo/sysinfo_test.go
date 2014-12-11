package sysinfo

import (
	"fmt"
	"testing"
)

func TestKernelVersion(t *testing.T) {
	p := NewParser(nil)
	v, err := p.KernelVersion()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: kernel version => '%s'\n", v)
}

func TestMajorMinorEqualOrGreaterThan(t *testing.T) {
	p := NewParser(nil)
	if p.KernelMajorMinorEqualOrGreaterThan("3.20.0-39-generic") {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

}
