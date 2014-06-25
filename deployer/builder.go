package deployer

type Builder interface {
	Run() (Artifact, error)
}
