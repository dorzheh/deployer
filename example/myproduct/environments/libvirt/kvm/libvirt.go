//
// FlowCreator interface implementation
//
package kvm

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common"
	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/common/bundle"
	"github.com/dorzheh/deployer/config/metadata"
	libvirtconf "github.com/dorzheh/deployer/config/metadata/libvirt/libvirt_kvm"
	"github.com/dorzheh/deployer/deployer"
	myprodcommon "github.com/dorzheh/deployer/example/myproduct/common"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt/libvirt_kvm"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/infra/comm/sshfs"
)

type FlowCreator struct {
	Filler              image.Rootfs
	SrcMetadataFile     string
	StorageConfigFile   string
	InputDataConfigFile string
	BundleConfigFile    string
	config              *metadata.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	var err error

	data := new(metadata.InputData)
	data.Lshw = filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw")
	data.InputDataConfigFile = c.InputDataConfigFile
	data.StorageConfigFile = c.StorageConfigFile
	data.BundleParser, err = bundle.NewParser(c.BundleConfigFile, new(bundle.DefaultBundle))
	if err != nil {
		return utils.FormatError(err)
	}
	if c.config, err = libvirtconf.CreateConfig(d, data); err != nil {
		return utils.FormatError(err)
	}

	return myprodcommon.NameToType(d.Ui, c.config.Bundle["name"].(string))
}

func (c *FlowCreator) CreateBuilders(d *deployer.CommonData) (b []deployer.Builder, err error) {
	if c.config.RemoteMode {
		d.Ui.Pb.IncreaseSleep("10s")
		d.Ui.Pb.DecreaseStep(4)
	}
	d.Ui.Pb.SetStep(15)

	var sshfsConf *sshfs.Config
	if c.config.RemoteMode {
		sshfsConf = &sshfs.Config{
			Common:      c.config.SshConfig,
			SshfsPath:   "",
			FusrmntPath: "",
		}
	}

	util := &image.Utils{
		Kpartx: filepath.Join(d.RootDir, "install", d.Arch, "bin/kpartx"),
	}

	for _, disk := range c.config.StorageConfig.Disks {
		imageData := &deployer.ImageBuilderData{
			ImageConfig: disk,
			RootfsMp:    d.RootfsMp,
			Filler:      c.Filler,
		}
		imageBuilder := &common.ImageBuilder{imageData, sshfsConf, util}
		b = append(b, imageBuilder)
	}

	metaData := &deployer.MetadataBuilderData{
		Source:   c.SrcMetadataFile,
		Dest:     c.config.DestMetadataFile,
		UserData: c.config.Metadata,
	}

	metadataBuilder := &common.MetadataBuilder{metaData, c.config.SshConfig}
	b = append(b, metadataBuilder)
	return
}

func (c *FlowCreator) CreatePostProcessor(d *deployer.CommonData) (p deployer.PostProcessor, err error) {
	d.Ui.Pb.SetSleep("2s")
	d.Ui.Pb.SetStep(10)
	p = libvirtpost.NewPostProcessor(c.config.SshConfig, false)
	return
}
