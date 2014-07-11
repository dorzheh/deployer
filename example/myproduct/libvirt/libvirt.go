//
// FlowCreator interface implementation
//
package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/builder/common/local"
	"github.com/dorzheh/deployer/builder/common/remote"
	libvirtconf "github.com/dorzheh/deployer/config/libvirt"
	"github.com/dorzheh/deployer/deployer"
	libvirtpost "github.com/dorzheh/deployer/post_processor/libvirt"
)

var networks = []string{"management", "data"}

type altdata struct {
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

	altd := &altdata{c.conf.Data, 2}
	metaData := &deployer.MetadataBuilderData{
		Source:   c.SrcMetadata,
		Dest:     c.conf.MetadataPath,
		UserData: altd,
	}

	var imageBuilder deployer.Builder
	if c.conf.Common.RemoteMode {
		imageBuilder = &remote.ImageBuilder{imageData,
			c.conf.Common.SshConfig,
			"path", "path",
		}
	} else {
		imageBuilder = &local.ImageBuilder{imageData,
			filepath.Join(c.conf.Common.Data.RootDir, "install",
				c.conf.Common.Data.Arch, "bin/lshw"), false}
	}

	var metadataBuilder deployer.Builder
	if c.conf.Common.RemoteMode {
		metadataBuilder = &remote.MetadataBuilder{metaData, c.conf.Common.SshConfig}
	} else {
		metadataBuilder = &local.MetadataBuilder{metaData, false}
	}

	b = append(b, imageBuilder, metadataBuilder)
	return
}

func (c *FlowCreator) CreateProvisioner() (p deployer.Provisioner, err error) {
	return
}

func (c *FlowCreator) CreatePostProcessor() (p deployer.PostProcessor, err error) {
	p = &libvirtpost.PostProcessor{
		Driver:      libvirtpost.NewDriver(c.conf.Common.SshConfig),
		DomName:     c.conf.Data.DomainName,
		StartDomain: false,
	}
	return
}
