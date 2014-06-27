package deployer

type Driver interface {
	DefineDomain(string) error
	StartDomain(string) error
	DestroyDomain(string) error
	UndefineDomain(string) error
	SetAutostart(string) error
	DomainExists(string) bool
	Emulator() (string, error)
}
