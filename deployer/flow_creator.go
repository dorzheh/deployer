package deployer

// FlowCreator is the main deployer interface.
type FlowCreator interface {
	// Creates appropriate config data needed for
	// creating builders,provisioners and post-processors.
	CreateConfig(*CommonData) error

	// Creates builders.
	CreateBuilders(*CommonData) ([]Builder, error)

	// Creates a post-processor.
	CreatePostProcessor(*CommonData) (PostProcessor, error)
}
