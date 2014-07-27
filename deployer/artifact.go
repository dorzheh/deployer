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

// Artifact is the interface to a real artifact implementation
// Any artifact object must implement this interface
type Artifact interface {
	// Artifact ID
	GetName() string

	// Path to artifact
	GetPath() string

	// Artifact type (either ImageArtifact or MetadataArtifact)
	GetType() ArtifactType

	// Destroys the artifact
	Destroy() error
}

// LocalArtifact implements Artifact interface
// and represents artifacts that creared and used locally
type CompressedArtifact struct {
	// Artifact name before compression
	RealName string

	// Path - full path to artifact
	Path string

	// Artifact type (either ImageArtifact or MetadataArtifact)
	Type ArtifactType

	SshConfig *ssh.Config
}

func (a *CompressedArtifact) GetName() string {
	return a.RealName
}

func (a *CompressedArtifact) GetPath() string {
	return a.Path
}

func (a *CompressedArtifact) GetType() ArtifactType {
	return a.Type
}

func (a *CompressedArtifact) Destroy() error {
	run := utils.RunFunc(a.SshConfig)
	if _, err := run("rm -f " + a.Path); err != nil {
		return err
	}
	return nil
}

type CommonArtifact struct {
	Name      string
	Path      string
	Type      ArtifactType
	SshConfig *ssh.Config
}

func (a *CommonArtifact) GetName() string {
	return a.Name
}

func (a *CommonArtifact) GetPath() string {
	return a.Path
}

func (a *CommonArtifact) GetType() ArtifactType {
	return a.Type
}

func (a *CommonArtifact) Destroy() error {
	run := utils.RunFunc(a.SshConfig)
	if _, err := run("rm -f " + a.Path); err != nil {
		return err
	}
	return nil
}
