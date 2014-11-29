package deployer

// PostProcessor is the interface that has to be
// implemented in order to post-process appropriate artifact.
type PostProcessor interface {
	// Processes given artifacts
	PostProcess([]Artifact) error
}
