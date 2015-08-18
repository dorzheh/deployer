//
// FlowCreator interface implementation
//
package kvm

import (
	"fmt"
	"path/filepath"

	"github.com/dorzheh/deployer/builder"
	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/metadata"
	libvirtconf "github.com/dorzheh/deployer/config/metadata/libvirt/libvirt_kvm"
	"github.com/dorzheh/deployer/deployer"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt/libvirt_kvm"
	"github.com/dorzheh/infra/comm/sshfs"
)

type FlowCreator struct {
	Filler              deployer.RootfsFiller
	StorageConfigFile   string
	InputDataConfigFile string
	BundleConfigFile    string
	SrcMetadataFile     string
	MetadataDir         string
	config              *metadata.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	var err error
	data := new(metadata.InputData)
	data.Lshw = filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw")
	data.InputDataConfigFile = c.InputDataConfigFile
	data.StorageConfigFile = c.StorageConfigFile
	data.TemplatesDir = c.MetadataDir

	if c.config, err = libvirtconf.CreateConfig(d, data); err != nil {
		return err
	}

	c.config.Ctrl.RegisterSteps(func(*FlowCreator, *deployer.CommonData) func() error {
		return func() error {
			fmt.Println("My step 1")
			return nil
		}
	}(c, d))

	c.config.Ctrl.RegisterSteps(func(*FlowCreator, *deployer.CommonData) func() error {
		return func() error {
			fmt.Println("My step 2")
			return nil
		}
	}(c, d))

	if err := c.config.Ctrl.RunSteps(); err != nil {
		return err
	}

	d.Ui.Pb.SetSleep("10s")
	d.Ui.Pb.SetStep(10)

	if c.config.RemoteMode {
		d.Ui.Pb.IncreaseSleep("5s")
		d.Ui.Pb.DecreaseStep(4)
	}
	return nil
}

func (c *FlowCreator) CreateBuilders(d *deployer.CommonData) (b []deployer.Builder, err error) {
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
		b = append(b, &builder.ImageBuilder{imageData, sshfsConf, util})
	}

	metaData := &deployer.MetadataBuilderData{
		Source:   c.SrcMetadataFile,
		Dest:     c.config.DestMetadataFile,
		UserData: c.config.Metadata,
	}

	b = append(b, &builder.MetadataBuilder{metaData, c.config.SshConfig})
	return
}

func (c *FlowCreator) CreatePostProcessor(d *deployer.CommonData) (p deployer.PostProcessor, err error) {
	d.Ui.Pb.SetSleep("1s")
	d.Ui.Pb.SetStep(50)
	p = libvirtpost.NewPostProcessor(c.config.SshConfig, false)
	return
}
