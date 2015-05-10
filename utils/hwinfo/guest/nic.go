package guest

import (
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

type PCI struct {
	// Domain address
	Domain string

	// Bus address
	Bus string

	// Slot address
	Slot string

	// Function address
	Function string
}

type NIC struct {
	Network string
	PCIAddr *PCI
	HostNIC *host.NIC
}

func NewNIC() *NIC {
	nic := new(NIC)
	nic.PCIAddr = new(PCI)
	return nic
}
