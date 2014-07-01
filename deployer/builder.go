package deployer

import "github.com/dorzheh/deployer/builder/common/image"

type Builder interface {
	Run() (Artifact, error)
}

type ImageBuilderData struct {
	ImagePath   string
	RootfsMp    string
	ImageConfig *image.Topology
	Filler      image.Rootfs
}

type MetadataBuilderData struct {
	Source   string
	Dest     string
	UserData interface{}
}
