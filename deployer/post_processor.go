package deployer

type PostProcessor interface {
	PostProcess([]Artifact) error
}
