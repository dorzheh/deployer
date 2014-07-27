package deployer

type PostProcessor interface {
	// Processes given artifacts
	PostProcess([]Artifact) error
}
