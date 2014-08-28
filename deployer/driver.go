package deployer

import (
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type Driver interface {
	// Creates appropriate domain
	DefineDomain(string) error

	// Starts appliance
	StartDomain(string) error

	// Stops appliance
	DestroyDomain(string) error

	// Removes appliance definitions(metadata)
	// and its references(doesn't remove appliance image)
	UndefineDomain(string) error

	// Sets appliance to start on host boot
	SetAutostart(string) error

	// Returns true if the given domain exists
	DomainExists(string) bool

	// Returns path to emulator(QEMU for example)
	Emulator(arch string) (string, error)
}

type HostinfoDriver interface {
	// Initialize the driver (mostly needed for UiGatherHWInfo)
	Init() error

	// Returns amount of installed RAM
	RAMSize() (uint, error)

	// Returns available CPUs
	CPUs() (uint, error)

	// Returns information related to the host's CPU
	CPUInfo() (*hwinfo.CPU, error)

	// Returns amount of NUMA nodes and appropriate CPUs per NUMA node
	NUMANodes() (map[uint][]uint, error)

	// Returns info related to the host's NICs
	NICs() ([]*hwinfo.NIC, error)
}
