package deployer

type Provisioner interface {
	Provision([]Artifact) ([]Artifact, error)
}
