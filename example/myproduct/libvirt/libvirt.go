//
// FlowCreator interface implementation
//
package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common"
	"github.com/dorzheh/deployer/builder/common/image"
	libvirtconf "github.com/dorzheh/deployer/config/libvirt"
	"github.com/dorzheh/deployer/deployer"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt"
	"github.com/dorzheh/infra/comm/sshfs"
)

type FlowCreator struct {
	Filler              image.Rootfs
	SrcMetadataFile     string
	TopologyConfigFile  string
	InputDataConfigFile string
	topology            *image.Topology
	config              *libvirtconf.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	var err error

	data := &libvirtconf.InputData{
		Lshw:             filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw"),
		InputDataXMLFile: c.InputDataConfigFile,
	}

	c.config, err = libvirtconf.CreateConfig(d, data)
	if err != nil {
		return err
	}
	f, err := image.ParseConfigFile(c.TopologyConfigFile)
	if err != nil {
		return err
	}
	c.topology, err = f.TypeToTopology(0)
	if err != nil {
		return err
	}
	return nil
}

func (c *FlowCreator) CreateBuilders(d *deployer.CommonData) (b []deployer.Builder, err error) {
	if err = d.Ui.Pb.SetSleep("10s"); err != nil {
		return
	}
	d.Ui.Pb.SetStep(15)

	imageData := &deployer.ImageBuilderData{
		ImagePath:   c.config.Metadata.ImagePath,
		RootfsMp:    d.RootfsMp,
		ImageConfig: c.topology,
		Filler:      c.Filler,
	}

	metaData := &deployer.MetadataBuilderData{
		Source:   c.SrcMetadataFile,
		Dest:     c.config.DestMetadataFile,
		UserData: c.config.Metadata,
	}
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

	imageBuilder := &common.ImageBuilder{imageData, sshfsConf, util}
	metadataBuilder := &common.MetadataBuilder{metaData, c.config.SshConfig}
	b = append(b, imageBuilder, metadataBuilder)
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
