// Represents any artifact crated by deployer
package deployer

import (
	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

type ArtifactType uint8

const (
	ImageArtifact ArtifactType = iota
	MetadataArtifact
)

// Artifact is the interface to a real artifact implementation.
// Any artifact object must implement this interface.
type Artifact interface {
	// Artifact ID.
	GetName() string

	// Path to artifact.
	GetPath() string

	// Artifact type (either ImageArtifact or MetadataArtifact).
	GetType() ArtifactType

	// Destroys the artifact.
	Destroy() error
}

// CommonArtifact represents an artifact properties.
type CommonArtifact struct {
	Name      string
	Path      string
	Type      ArtifactType
	SshConfig *ssh.Config
}

// GetName returns artifact's name.
func (a *CommonArtifact) GetName() string {
	return a.Name
}

// GetPath returns artifact's path.
func (a *CommonArtifact) GetPath() string {
	return a.Path
}

// GetType returns artifact's type (metadata or image).
func (a *CommonArtifact) GetType() ArtifactType {
	return a.Type
}

// Destroy is responsible for removing appropriate artifact.
func (a *CommonArtifact) Destroy() error {
	run := utils.RunFunc(a.SshConfig)
	if _, err := run("rm -f " + a.Path); err != nil {
		return utils.FormatError(err)
	}
	return nil
}
