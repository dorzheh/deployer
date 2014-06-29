package remote

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/gssh/sshfs"
	"github.com/dorzheh/infra/utils"
)

type ImageBuilder struct {
	SshConfig           *ssh.Config
	ImagePath           string
	BuildScriptPath     string
	RootfsMp            string
	BuildScriptUserData interface{}
	ImageConfig         *image.Topology
	Filler              image.Rootfs
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	f, err := ioutil.ReadFile(b.BuildScriptPath)
	if err != nil {
		return nil, err
	}
	script, err := deployer.ProcessTemplate(string(f), b.BuildScriptUserData)
	if err != nil {
		return nil, err
	}

	run := deployer.RunFunc(b.SshConfig)
	cmd := fmt.Sprintf("echo %s > /tmp/run_script;sudo bash -x /tmp/run_script;rm -f  /tmp/run_script", script)
	if _, err := run(cmd); err != nil {
		return nil, err
	}

	if err := utils.CreateDirRecursively(b.RootfsMp, 0755, 0, 0, false); err != nil {
		return nil, err
	}
	defer os.RemoveAll(b.RootfsMp)

	conf := &sshfs.Config{
		Common:      b.SshConfig,
		SshfsPath:   "",
		FusrmntPath: "",
	}

	s, err := sshfs.NewClient(conf)
	if err != nil {
		return nil, err
	}
	if err := s.Attach("/dummy", b.RootfsMp); err != nil {
		return nil, err
	}
	defer s.Detach(b.RootfsMp)

	if b.Filler != nil {
		if err := b.Filler.MakeRootfs(b.RootfsMp); err != nil {
			return nil, err
		}
		// install application.
		if err := b.Filler.InstallApp(b.RootfsMp); err != nil {
			return nil, err
		}
	}
	return &deployer.RemoteArtifact{
		Name: b.ImageConfig.Name,
		Path: b.ImagePath,
		Type: deployer.ImageArtifact,
	}, nil
}

type MetadataBuilder struct {
	SshConfig *ssh.Config
	Source    string
	Dest      string
	UserData  interface{}
}

func (b *MetadataBuilder) Run() (deployer.Artifact, error) {
	f, err := ioutil.ReadFile(b.Source)
	if err != nil {
		return nil, err
	}
	data, err := deployer.ProcessTemplate(string(f), b.UserData)
	if err != nil {
		return nil, err
	}

	run := deployer.RunFunc(b.SshConfig)
	cmd := fmt.Sprintf("echo %s > %s", data, b.Dest)
	if _, err := run(cmd); err != nil {
		return nil, err
	}
	return &deployer.RemoteArtifact{
		Name: filepath.Base(b.Dest),
		Path: b.Dest,
		Type: deployer.MetadataArtifact,
	}, nil
}
