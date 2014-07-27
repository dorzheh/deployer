package deployer

import (
	"os"
	"testing"

	ssh "github.com/dorzheh/infra/comm/common"
)

const artifactPath = "/tmp/testArtifact"

func TestCreateDestroyLocalCommonArtifact(t *testing.T) {
	if _, err := os.Create(artifactPath); err != nil {
		t.Error(err)
		return
	}
	a := &CommonArtifact{
		Name: "testArtifact",
		Path: artifactPath,
		Type: ImageArtifact,
	}
	if err := a.Destroy(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDestroyRemoteCommonArtifact(t *testing.T) {
	if _, err := os.Create(artifactPath); err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(artifactPath)

	conf := &ssh.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "root",
		Password:    "d",
		PrvtKeyFile: "",
	}
	a := &CommonArtifact{
		Name:      "testArtifact",
		Path:      artifactPath,
		Type:      ImageArtifact,
		SshConfig: conf,
	}
	if err := a.Destroy(); err != nil {
		t.Fatal(err)
	}
}
