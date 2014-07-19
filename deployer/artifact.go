// Represents any artifact crated by deployer
package deployer

import "os"

type ArtifactType uint8

const (
	ImageArtifact ArtifactType = iota
	MetadataArtifact
)

// Artifact is the interface to a real artifact implementation
// Any artifact object must implement this interface
type Artifact interface {
	// artifact ID
	GetName() string

	// path to artifact
	GetPath() string

	// artifact type (either ImageArtifact or MetadataArtifact)
	GetType() ArtifactType

	// destroys the artifact
	Destroy() error
}

// LocalArtifact implements Artifact interface
// and represents artifacts that creared and used locally
type LocalArtifact struct {
	// Name - artifact name
	Name string

	// Path - full path to artifact
	Path string

	// artifact type (either ImageArtifact or MetadataArtifact)
	Type ArtifactType
}

func (a *LocalArtifact) GetName() string {
	return a.Name
}

func (a *LocalArtifact) GetPath() string {
	return a.Path
}

func (a *LocalArtifact) GetType() ArtifactType {
	return a.Type
}

func (a *LocalArtifact) Destroy() error {
	return os.Remove(a.Path)
}

type RemoteArtifact struct {
	Name string
	Path string
	Type ArtifactType
}

func (a *RemoteArtifact) GetName() string {
	return a.Name
}

func (a *RemoteArtifact) GetPath() string {
	return a.Path
}

func (a *RemoteArtifact) GetType() ArtifactType {
	return a.Type
}

func (a *RemoteArtifact) Destroy() error {
	return nil
}
