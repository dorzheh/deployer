package deployer

type Provisioner interface {
	// Provisions given artifacts
	Provision([]Artifact) ([]Artifact, error)
}
