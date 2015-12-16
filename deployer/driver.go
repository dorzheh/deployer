package deployer

import (
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

// EnvDriver is the interface that has to be implemented in order
// to communicate with VM over API belonging to appropriate environment
type EnvDriver interface {
	// Returns the driver identificator
	Id() string

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

	// Returns maximal Virtual CPUs per guest
	MaxVCPUsPerGuest() int

	// Returns true if all Virtual Appliances in the farm configured for one-to-one CPU pinning
	AllCPUsPinned() (bool, error)
}

// HostinfoDriver is the interface that has to be implemented in order to
// gather appropriate HW information from either local or remote host
type HostinfoDriver interface {
	// Initialize the driver
	Init() error

	// Returns amount of installed RAM.
	RAMSize() (int, error)

	// Returns available CPUs.
	CPUs() (int, error)

	// Returns information related to the host's CPU.
	CPUInfo() (*host.CPU, error)

	// Returns information about installed NUMA nodes
	NUMAInfo() (host.NUMANodes, error)

	// Returns info related to the host's NICs.
	NICs() (host.NICList, error)
}
