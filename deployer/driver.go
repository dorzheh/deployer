package deployer

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
	Emulator() (string, error)
}
