package deployer

import (
	"github.com/dorzheh/deployer/builder/image"
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
// needed by appropriate image builder.
type ImageBuilderData struct {
	// ImageConfig - XML metadata containing image topology configuration.
	ImageConfig *image.Disk

	// Filler - implementation of deployer.RootfsFiller interface.
	Filler RootfsFiller

	// RootfsMp - path to the mount point where the image
	// artifact will be mounted during customization.
	RootfsMp string
}

// MetadataBuilderData represents the common data
// needed by appropriate metadata builder.
type MetadataBuilderData struct {
	// Source - path to a source metadata artifact.
	Source string

	// Dest - path to destination metadata artifact.
	Dest string

	// UserData - any data provided by user and that will be
	// written to destination metadata.
	UserData interface{}
}

// DirBuilderData represents the common data
// needed by appropriate image builder.
type DirBuilderData struct {
	// Filler - implementation of RootfsFiller interface.
	Filler RootfsFiller

	// RootfsPath - path to rootfs
	RootfsPath string
}
