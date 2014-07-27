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
type ImageBuilderData struct {
	// ImagePath - path to the image artifact
	ImagePath string
	// RootfsMp - path to the mount point where the image
	// artifact will be mounted during customization
	RootfsMp string
	// ImageConfig - XML metadata containing image topology configuration
	ImageConfig *image.Topology
	// Filler - implementation of image.Rootfs interface
	Filler image.Rootfs
}

// MetadataBuilderData represents the common data
// needed by appropriate metadata builder
type MetadataBuilderData struct {
	// Source - path to a source metadata artifact
	Source string
	// Dest - path to destination metadata artifact
	Dest string
	// UserData - any data provided by user and that will be
	// written to destination metadata
	UserData interface{}
}
