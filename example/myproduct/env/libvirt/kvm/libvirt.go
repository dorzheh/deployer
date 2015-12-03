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
	"github.com/dorzheh/deployer/controller"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct/common"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt/libvirt_kvm"
	"github.com/dorzheh/infra/comm/sshfs"
)

var mainConfig = map[string]string{
	"config_dir":             "comp/env/libvirt/kvm/config",
	"metadata_dir":           "comp/env/libvirt/kvm/metadata",
	"metadata_file":          "comp/env/libvirt/kvm/metadata/domain.tmplt.xml",
	"storage_config_file":    "comp/env/common/config/storage_config.xml",
	"input_data_config_file": "comp/env/libvirt/kvm/config/input_data_config.xml",
}

type FlowCreator struct {
	config *metadata.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	d.DefaultExportDir = "/var/lib/libvirt/images"
	data := new(metadata.InputData)
	data.Lshw = filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw")
	data.InputDataConfigFile = filepath.Join(d.RootDir, mainConfig["input_data_config_file"])
	data.StorageConfigFile = filepath.Join(d.RootDir, mainConfig["storage_config_file"])
	data.TemplatesDir = filepath.Join(d.RootDir, mainConfig["metadata_dir"])

	var err error
	if c.config, err = libvirtconf.CreateConfig(d, data); err != nil {
		return err
	}

	controller.RegisterSteps(func() func() error {
		return func() error {
			fmt.Println("My step 1")
			return nil
		}
	}())

	controller.RegisterSteps(func() func() error {
		return func() error {
			fmt.Println("My step 2")
			return nil
		}
	}())

	if err := controller.RunSteps(); err != nil {
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
	for _, disk := range c.config.StorageConfig.Configs[0].Disks {
		imageData := &deployer.ImageBuilderData{
			ImageConfig: disk,
			RootfsMp:    d.RootfsMp,
			Filler:      common.ImageFiller(d, mainConfig["config_dir"]),
		}
		b = append(b, &builder.ImageBuilder{imageData, sshfsConf, util})
	}

	metaData := &deployer.MetadataBuilderData{
		Source:   filepath.Join(d.RootDir, mainConfig["metadata_file"]),
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
