//
// FlowCreator interface implementation
//
package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common"
	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/metadata"
	libvirtconf "github.com/dorzheh/deployer/config/metadata/libvirt"
	"github.com/dorzheh/deployer/deployer"
	myprodcommon "github.com/dorzheh/deployer/example/myproduct/common"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt"
	"github.com/dorzheh/infra/comm/sshfs"
)

type FlowCreator struct {
	Filler              image.Rootfs
	SrcMetadataFile     string
	StorageConfigFile   string
	InputDataConfigFile string
	config              *metadata.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	data := &metadata.InputData{
		Lshw:                filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw"),
		InputDataConfigFile: c.InputDataConfigFile,
		StorageConfigFile:   c.StorageConfigFile,
	}

	var err error
	if c.config, err = libvirtconf.CreateConfig(d, data); err != nil {
		return err
	}

	return myprodcommon.NameToType(d.Ui, c.config.Metadata.DomainName)
}

func (c *FlowCreator) CreateBuilders(d *deployer.CommonData) (b []deployer.Builder, err error) {
	if err = d.Ui.Pb.SetSleep("10s"); err != nil {
		return
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
		Grub:   filepath.Join(d.RootDir, "install", d.Arch, "bin/grub"),
		Kpartx: filepath.Join(d.RootDir, "install", d.Arch, "bin/kpartx"),
	}

	for _, diskconf := range c.config.StorageConfig.Disks {
		imageData := &deployer.ImageBuilderData{
			ImageConfig: diskconf,
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
	if err = d.Ui.Pb.SetSleep("3s"); err != nil {
		return
	}
	d.Ui.Pb.SetStep(10)

	p = &libvirtpost.PostProcessor{
		Driver:      libvirtpost.NewDriver(c.config.SshConfig),
		StartDomain: false,
	}
	return
}
