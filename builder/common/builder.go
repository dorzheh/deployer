// Local builder is responsible for creating local artifacts
package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

// ImageBuilder represents properties related to a local image builder
type ImageBuilder struct {
	// deployer.ImageBuilderData represents common data
	*deployer.ImageBuilderData

	// image.RemoteConfig represents remote configuration
	// facility needed by the builder
	Rconf *image.RemoteConfig

	// GrubPath - path to grub 1.x binary
	GrubPath string

	// Compress indicates if the artifact should be compressed
	Compress bool
}

func (b *ImageBuilder) Id() string {
	if b.Rconf == nil {
		return "RemoteImageBuilder"
	}
	return "LocalImageBuilder"
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	if err := os.MkdirAll(b.RootfsMp, 0755); err != nil {
		return nil, err
	}
	defer func() error {
		return os.RemoveAll(b.RootfsMp)
	}()
	// create new image artifact
	img, err := image.New(b.ImagePath, b.RootfsMp, b.ImageConfig, b.Rconf)
	if err != nil {
		return nil, err
	}
	// interrupt handler
	img.ReleaseOnInterrupt()
	defer func() error {
		if err := img.Release(); err != nil {
			return err
		}
		if b.ImageConfig.Bootable {
			return img.MakeBootable(b.GrubPath)
		}
		return nil
	}()
	// parse the image
	if err := img.Parse(); err != nil {
		return nil, err
	}
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

	var a deployer.Artifact
	if b.Compress {
		origName := filepath.Base(b.ImagePath)
		newImagePath, err := compressArtifact(b.ImagePath)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(b.ImagePath); err != nil {
			return nil, err
		}
		b.ImagePath = newImagePath

		a = &deployer.CompressedArtifact{
			RealName: origName,
			Path:     b.ImagePath,
			Type:     deployer.ImageArtifact,
		}
	} else {
		a = &deployer.CommonArtifact{
			Name: b.ImageConfig.Name,
			Path: b.ImagePath,
			Type: deployer.ImageArtifact,
		}
	}
	return a, nil
}

// MetadataBuilder represents properties related to a local metadata builder
type MetadataBuilder struct {
	// *deployer.MetadataBuilderData represents common data
	*deployer.MetadataBuilderData

	// SshCpnfig provides appropriate properties
	// for being able to create remote connection
	SshConfig *ssh.Config
}

func (b *MetadataBuilder) Id() string {
	return "MetadataBuilder"
}

func (b *MetadataBuilder) Run() (deployer.Artifact, error) {
	f, err := ioutil.ReadFile(b.Source)
	if err != nil {
		return nil, err
	}
	data, err := utils.ProcessTemplate(string(f), b.UserData)
	if err != nil {
		return nil, err
	}

	run := utils.RunFunc(b.SshConfig)
	if _, err := run(fmt.Sprintf("echo \"%s\" > %s", data, b.Dest)); err != nil {
		return nil, err
	}
	return &deployer.CommonArtifact{
		Name: filepath.Base(b.Dest),
		Path: b.Dest,
		Type: deployer.MetadataArtifact,
	}, nil
}

// InstanceBuilder represents properties related to a local instance builder
// The common usage of InstanceBuiler: running deployer on a cloud instance
type InstanceBuilder struct {
	Filler image.Rootfs
}

func (b *InstanceBuilder) Run() (a deployer.Artifact, err error) {
	if err = b.Filler.MakeRootfs("/"); err != nil {
		return
	}
	if err = b.Filler.InstallApp("/"); err != nil {
		return
	}
	return
}

// compressArtifact is intended for compressing artifacts in case of necessity
func compressArtifact(path string) (string, error) {
	dir := filepath.Dir(path)
	oldArtifactFile := filepath.Base(path)
	if err := os.Chdir(dir); err != nil {
		return "", err
	}
	newArtifactFile := oldArtifactFile + ".tgz"
	var stderr bytes.Buffer
	cmd := exec.Command("tar", "cfzp", newArtifactFile, oldArtifactFile)
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("%s [%s]", stderr.String(), err)
	}
	return filepath.Join(dir, newArtifactFile), nil
}
