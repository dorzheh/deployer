package remote

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/comm/ssh"
	"github.com/dorzheh/infra/utils"
)

type ImageBuilder struct {
	ConnFunc            func(*ssh.Config) (*ssh.SshConn, error)
	SshConfig           *ssh.Config
	ImagePath           string
	BuildScriptPath     string
	RootfsMp            string
	BuildScriptUserData interface{}
	ImageConfig         *image.Topology
	Filler              image.Rootfs
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	sshfs, err := exec.LookPath("sshfs")
	if err != nil {
		return nil, err
	}
	fusermount, err := exec.LookPath("fusermount")
	if err != nil {
		return nil, err
	}

	println(sshfs, fusermount)

	conn, err := b.ConnFunc(b.SshConfig)
	if err != nil {
		return nil, err
	}
	defer conn.ConnClose()

	f, err := ioutil.ReadFile(b.BuildScriptPath)
	if err != nil {
		return nil, err
	}
	script, err := deployer.ProcessTemplate(string(f), b.BuildScriptUserData)
	if err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf("echo %s > /tmp/run_script;sudo bash -x /tmp/run_script;rm -f  /tmp/run_script", script)
	if _, stderr, err := conn.Run(cmd); err != nil {
		return nil, fmt.Errorf("%s [%s]", stderr, err)
	}

	if err := utils.CreateDirRecursively(b.RootfsMp, 0755, 0, 0, false); err != nil {
		return nil, err
	}
	defer os.RemoveAll(b.RootfsMp)

	//argsPat := fmt.Sprintf("%s@%s:%s %s -o %s,idmap=user,compression=no,nonempty,Ciphers=arcfour",
	//	s.account.name, s.ip, remoteShare, localMount, mo)
	//if out, err := exec.Command("sshfs", strings.Fields(argsPat)...).CombinedOutput(); err != nil {
	//	return fmt.Errorf("%s:%s", out, err)
	//}
	//if out, err := exec.Command("fusermount", "-u", sh.MountPoint).CombinedOutput(); err != nil {
	//                     return fmt.Errorf("%s:%s", out, err)
	//             }

	//}()
	// create and customize rootfs
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
	ConnFunc  func(*ssh.Config) (*ssh.SshConn, error)
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
	conn, err := b.ConnFunc(b.SshConfig)
	if err != nil {
		return nil, err
	}
	cmd := fmt.Sprintf("echo %s > %s", data, b.Dest)
	if _, stderr, err := conn.Run(cmd); err != nil {
		return nil, fmt.Errorf("%s [%s]", stderr, err)
	}
	return &deployer.RemoteArtifact{
		Name: filepath.Base(b.Dest),
		Path: b.Dest,
		Type: deployer.MetadataArtifact,
	}, nil
}

// sshfs me@x.x.x.x:/remote/path /local/path/ -o IdentityFile=/path/to/key
