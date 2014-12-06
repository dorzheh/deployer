package deployer

import (
	"github.com/dorzheh/deployer/utils/hwinfo"
)

// Driver is the interface that has to be implemented in order
// to communicate with our appliance over appropriate API
type Driver interface {
	// Creates appropriate domain.
	DefineDomain(string) error

	// Starts appliance.
	StartDomain(string) error

	// Stops appliance
	DestroyDomain(string) error

	// Removes appliance definitions(metadata)
	// and its references(doesn't remove appliance image).
	UndefineDomain(string) error

	// Sets appliance to start on host boot.
	SetAutostart(string) error

	// Returns true if the given domain exists.
	DomainExists(string) bool

	// Returns path to emulator(QEMU for example).
	Emulator(arch string) (string, error)

	// Returns driver version (for example if the driver is libvirt the function
	// will return libvirt API version)
	Version() (string, error)
}

// HostinfoDriver is the interface that has to be implemented in order to
// gather appropriate HW information from either local or remote host
type HostinfoDriver interface {
	// Initialize the driver (mostly needed for UiGatherHWInfo).
	Init() error

	// Returns amount of installed RAM.
	RAMSize() (uint, error)

	// Returns available CPUs.
	CPUs() (uint, error)

	// Returns information related to the host's CPU.
	CPUInfo() (*hwinfo.CPU, error)

	// Returns amount of NUMA nodes and appropriate CPUs per NUMA node.
	NUMANodes() (map[uint][]uint, error)

	// Returns info related to the host's NICs.
	NICs() ([]*hwinfo.NIC, error)
}
