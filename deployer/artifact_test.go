package deployer

import (
	"os"
	"testing"
)

const artifactPath = "/tmp/testArtifact"

func TestCreateDestroyLocalArtifact(t *testing.T) {
	if _, err := os.Create(artifactPath); err != nil {
		t.Error(err)
		return
	}
	a := &LocalArtifact{
		Name: "testArtifact",
		Path: artifactPath,
		Type: ImageArtifact,
	}
	if err := a.Destroy(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDestroyRemoteArtifact(t *testing.T) {
	if _, err := os.Create(artifactPath); err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(artifactPath)

	a := &RemoteArtifact{
		Name: "testArtifact",
		Path: artifactPath,
		Type: ImageArtifact,
	}
	if err := a.Destroy(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(artifactPath); err != nil {
		t.Fatalf("RemoteArtifact shouldn't be destroyed")
	}
}
