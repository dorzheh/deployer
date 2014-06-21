package deployer

type PostProcessor interface {
	Process([]Artifact) error
}
