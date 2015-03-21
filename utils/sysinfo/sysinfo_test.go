package sysinfo

import (
	"fmt"
	"testing"
)

func TestKernelVersion(t *testing.T) {
	c := NewCollector(nil)
	v, err := c.KernelVersion()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: kernel version => '%s'\n", v)
}

func TestMajorMinorEqualOrGreaterThan(t *testing.T) {
	c := NewCollector(nil)
	if c.KernelMajorMinorEqualOrGreaterThan("3.20.0-39-generic") {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

}
