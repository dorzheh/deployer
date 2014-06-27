package deployer

type FlowCreator interface {
	CreateConfig(*CommonData) error
	CreateBuilders() ([]Builder, error)
	CreateProvisioner() (Provisioner, error)
	CreatePostProcessor() (PostProcessor, error)
}
