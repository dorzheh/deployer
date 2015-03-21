package sysinfo

// This package provides required methods gathering system information.
// For example, comparing kernel version if  multiqueue support
// in virtio network driver is required.

import (
	"fmt"

	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

type Collector struct {
	run func(string) (string, error)
}

func NewCollector(config *ssh.Config) *Collector {
	c := new(Collector)
	c.run = utils.RunFunc(config)
	return c
}

func (c *Collector) KernelVersion() (string, error) {
	return c.run("uname -r")
}

func (c *Collector) KernelMajorMinorEqualOrGreaterThan(other string) bool {
	current, err := c.KernelVersion()
	if err != nil {
		return false
	}

	var curMaj int
	var curMin int
	fmt.Sscanf(current, "%d.%d", &curMaj, &curMin)

	var otherMaj int
	var otherMin int
	fmt.Sscanf(other, "%d.%d", &otherMaj, &otherMin)

	switch {
	case curMaj > otherMaj:
		return true
	case curMaj == otherMaj:
		if curMin >= otherMin {
			return true
		}
	}
	return false
}
