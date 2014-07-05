package local

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/utils"
)

type ImageBuilder struct {
	*deployer.ImageBuilderData
	GrubPath string
	Compress bool
}

func (b *ImageBuilder) Id() string {
	return "LocalImageBuilder"
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	if err := utils.CreateDirRecursively(b.RootfsMp, 0755, 0, 0, false); err != nil {
		return nil, err
	}
	defer os.RemoveAll(b.RootfsMp)
	// create new image object
	img, err := image.New(b.ImagePath, b.RootfsMp, b.ImageConfig)
	if err != nil {
		return nil, err
	}
	// parse the image
	if err := img.Parse(); err != nil {
		return nil, err
	}
	// interrupt handler
	img.ReleaseOnInterrupt()
	defer func() {
		if err := img.Release(); err != nil {
			panic(err)
		}
		if b.ImageConfig.Bootable {
			if err := img.MakeBootable(b.GrubPath); err != nil {
				panic(err)
			}
		}
	}()
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

	origName := filepath.Base(b.ImagePath)
	if b.Compress {
		newImagePath, err := compressArtifact(b.ImagePath)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(b.ImagePath); err != nil {
			return nil, err
		}
		b.ImagePath = newImagePath
	}
	return &deployer.LocalArtifact{
		Name: origName,
		Path: b.ImagePath,
		Type: deployer.ImageArtifact,
	}, nil
}

type MetadataBuilder struct {
	*deployer.MetadataBuilderData
	Compress bool
}

func (b *MetadataBuilder) Id() string {
	return "LocalMetadataBuilder"
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
	if err := ioutil.WriteFile(b.Dest, data, 0644); err != nil {
		return nil, err
	}

	origName := filepath.Base(b.Dest)
	if b.Compress {
		newDest, err := compressArtifact(b.Dest)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(b.Dest); err != nil {
			return nil, err
		}
		b.Dest = newDest
	}
	return &deployer.LocalArtifact{
		Name: origName,
		Path: b.Dest,
		Type: deployer.MetadataArtifact,
	}, nil
}

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
