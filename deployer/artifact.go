package deployer

import "os"

type ArtifactType uint8

const (
	ImageArtifact ArtifactType = iota
	MetadataArtifact
)

type Artifact interface {
	GetName() string
	GetPath() string
	GetType() ArtifactType
	Destroy() error
}

type LocalArtifact struct {
	Name string
	Path string
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
