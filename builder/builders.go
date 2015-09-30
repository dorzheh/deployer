package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/sshfs"
)

// ImageBuilder represents properties related to a local image builder
type ImageBuilder struct {
	// deployer.ImageBuilderData represents common data
	*deployer.ImageBuilderData

	// SshfsConfig represents remote configuration
	// facility needed by the builder
	SshfsConfig *sshfs.Config

	// set of utilities needed for image manipulation
	Utils *image.Utils
}

func (b *ImageBuilder) Id() string {
	if b.SshfsConfig == nil {
		return "LocalImageBuilder"
	}
	return "RemoteImageBuilder"
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	if err := os.MkdirAll(b.RootfsMp, 0755); err != nil {
		return nil, utils.FormatError(err)
	}

	defer os.RemoveAll(b.RootfsMp)

	// create new image artifact
	img, err := image.New(b.ImageConfig, b.RootfsMp, b.Utils, b.SshfsConfig)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	// interrupt handler
	img.ReleaseOnInterrupt()
	defer func() {
		img.Cleanup()
	}()

	// parse the image
	if err := img.Parse(); err != nil {
		return nil, utils.FormatError(err)
	}
	// customize rootfs
	if b.Filler != nil {
		if err := b.Filler.CustomizeRootfs(b.RootfsMp); err != nil {
			return nil, utils.FormatError(err)
		}
		// install application.
		if err := b.Filler.InstallApp(b.RootfsMp); err != nil {
			return nil, utils.FormatError(err)
		}
	}
	if b.ImageConfig.Bootable {
		if err := img.MakeBootable(); err != nil {
			return nil, utils.FormatError(err)
		}
	}
	if b.Filler != nil {
		if err := b.Filler.RunHooks(b.RootfsMp); err != nil {
			return nil, utils.FormatError(err)
		}
	}
	if err := img.Cleanup(); err != nil {
		return nil, utils.FormatError(err)
	}
	if err := img.Convert(); err != nil {
		return nil, utils.FormatError(err)
	}
	return &deployer.CommonArtifact{
		Name: filepath.Base(b.ImageConfig.Path),
		Path: b.ImageConfig.Path,
		Type: deployer.ImageArtifact,
	}, nil
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
	// in case no source template exists apparently we should use the default metadata
	_, err := os.Stat(b.Source)
	if err != nil {
		b.Source = b.Dest
	}

	f, err := ioutil.ReadFile(b.Source)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	data, err := utils.ProcessTemplate(string(f), b.UserData)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	run := utils.RunFunc(b.SshConfig)
	if _, err := run(fmt.Sprintf("echo \"%s\" > %s", data, b.Dest)); err != nil {
		return nil, utils.FormatError(err)
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
	Filler deployer.RootfsFiller
}

func (b *InstanceBuilder) Id() string {
	return "InstanceBuilder"
}

func (b *InstanceBuilder) Run() (a deployer.Artifact, err error) {
	if err = b.Filler.CustomizeRootfs("/"); err != nil {
		err = utils.FormatError(err)
		return
	}
	if err = b.Filler.InstallApp("/"); err != nil {
		err = utils.FormatError(err)
		return
	}
	return
}

// DirBuilder is intended for building directory backed appliances(LXC,OpenVZ and so forth)
type DirBuilder struct {
	// deployer.DirBuilderData represents common data
	*deployer.DirBuilderData

	// SshfsConfig represents remote configuration
	// facility needed by the builder
	SshfsConfig *sshfs.Config
}

func (b *DirBuilder) Id() string {
	if b.SshfsConfig == nil {
		return "LocalDirBuilder"
	}
	return "RemoteDirBuilder"
}

func (b *DirBuilder) Run() (a deployer.Artifact, err error) {
	// customize rootfs
	if b.Filler != nil {
		if err = b.Filler.CustomizeRootfs(b.RootfsPath); err != nil {
			err = utils.FormatError(err)
			return
		}
		// install application
		if err = b.Filler.InstallApp(b.RootfsPath); err != nil {
			err = utils.FormatError(err)
			return
		}
	}
	return
}
