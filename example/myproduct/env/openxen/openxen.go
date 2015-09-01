package openxen

import (
	"fmt"
	"path/filepath"

	"github.com/dorzheh/deployer/builder"
	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/metadata"
	xenconf "github.com/dorzheh/deployer/config/metadata/openxen/xen_xl"
	"github.com/dorzheh/deployer/controller"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct/common"
	xenpost "github.com/dorzheh/deployer/post_processor/openxen/xen_xl"
	"github.com/dorzheh/infra/comm/sshfs"
)

var mainConfig = map[string]string{
	"config_dir":             "comp/env/openxen/config",
	"metadata_file":          "comp/env/openxen/metadata/domu.pvhvm.tmplt.cfg",
	"storage_config_file":    "comp/env/common/config/storage_config.xml",
	"input_data_config_file": "comp/env/openxen/config/input_data_config.xml",
}

type FlowCreator struct {
	config *metadata.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	d.DefaultExportDir = "/var/lib/xen"
	data := new(metadata.InputData)
	data.Lshw = filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw")
	data.InputDataConfigFile = filepath.Join(d.RootDir, mainConfig["input_data_config_file"])
	data.StorageConfigFile = filepath.Join(d.RootDir, mainConfig["storage_config_file"])

	var err error
	if c.config, err = xenconf.CreateConfig(d, data); err != nil {
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

	// Xen XL metadata requires that the RAM size will be represented in Megabytes
	c.config.Metadata.RAM /= 1024

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
	p = xenpost.NewPostProcessor(c.config.SshConfig, true)
	return
}
