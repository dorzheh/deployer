package deployer

type FlowCreator interface {
	// Creates appropriate config data needed for
	// creating builders,provisioners and post-processors
	CreateConfig(*CommonData) error

	// Creates builders
	CreateBuilders() ([]Builder, error)

	// Creates a post-processor
	CreatePostProcessor() (PostProcessor, error)
}
