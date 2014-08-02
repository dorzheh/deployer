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

var networks = []string{"management", "data"}

type prodData struct {
	*libvirtconf.CommonMetadata
	Cpus int
}

type FlowCreator struct {
	Filler      image.Rootfs
	SrcMetadata string
	topology    *image.Topology
	conf        *libvirtconf.Config
}

func (c *FlowCreator) CreateConfig(d *deployer.CommonData) error {
	var err error

	data := &libvirtconf.InputData{
		Networks: networks,
		LshwPath: filepath.Join(d.RootDir, "install", d.Arch, "bin/lshw"),
	}
	c.conf, err = libvirtconf.CreateConfig(d, data)
	if err != nil {
		return err
	}
	f, err := image.ParseConfigFile(filepath.Join(c.conf.Common.Data.RootDir, "env/common/topology.xml"))
	if err != nil {
		return err
	}
	c.topology, err = f.TypeToTopology(0)
	if err != nil {
		return err
	}
	return nil
}

func (c *FlowCreator) CreateBuilders() (b []deployer.Builder, err error) {
	imageData := &deployer.ImageBuilderData{
		ImagePath:   c.conf.Data.ImagePath,
		RootfsMp:    c.conf.Common.Data.RootfsMp,
		ImageConfig: c.topology,
		Filler:      c.Filler,
	}

	data := &prodData{c.conf.Data, 2}
	metaData := &deployer.MetadataBuilderData{
		Source:   c.SrcMetadata,
		Dest:     c.conf.MetadataPath,
		UserData: data,
	}
	var sshfsConf *sshfs.Config
	if c.conf.Common.RemoteMode {
		sshfsConf = &sshfs.Config{
			Common:      c.conf.Common.SshConfig,
			SshfsPath:   "",
			FusrmntPath: "",
		}
	}

	u := &image.Utils{
		Grub: filepath.Join(c.conf.Common.Data.RootDir, "install",
			c.conf.Common.Data.Arch, "bin/grub"),
		Kpartx: filepath.Join(c.conf.Common.Data.RootDir, "install",
			c.conf.Common.Data.Arch, "bin/kpartx"),
	}

	imageBuilder := &common.ImageBuilder{imageData, sshfsConf, u, false}
	metadataBuilder := &common.MetadataBuilder{metaData, c.conf.Common.SshConfig}
	b = append(b, imageBuilder, metadataBuilder)
	return
}

func (c *FlowCreator) CreateProvisioner() (p deployer.Provisioner, err error) {
	return
}

func (c *FlowCreator) CreatePostProcessor() (p deployer.PostProcessor, err error) {
	p = &libvirtpost.PostProcessor{
		Driver:      libvirtpost.NewDriver(c.conf.Common.SshConfig),
		StartDomain: false,
	}
	return
}
