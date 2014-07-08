package deployer

import (
	"github.com/dorzheh/deployer/builder/common/image"
)

// Implementers of the interface are responsible for creating
// appropriate artifact.
type Builder interface {
	// Id of the build
	Id() string
	// Run the build
	Run() (Artifact, error)
}

// ImageBuilderData represents the common data
// needed by appropriate image builder
// ImagePath - path to the image artifact
// RootfsMp - path to the mount point where the image
// artifact will be mounted during customization
// ImageConfig - XML metadata containing image topology configuration
// Filler - implementation of image.Rootfs interface
type ImageBuilderData struct {
	ImagePath   string
	RootfsMp    string
	ImageConfig *image.Topology
	Filler      image.Rootfs
}

// MetadataBuilderData represents the common data
// needed by appropriate metadata builder
// Source - path to a source metadata artifact
// Dest - path to destination metadata artifact
// UserData - any data provided by user and that will be
// written to destination metadata
type MetadataBuilderData struct {
	Source   string
	Dest     string
	UserData interface{}
}
