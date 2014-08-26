package myproduct

import (
	"path/filepath"

	main "github.com/dorzheh/deployer"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct/libvirt"
	"github.com/dorzheh/infra/utils/archutils"
)

const (
	LIBVIRT = "1"
	DUMMY2  = "2"
	DUMMY3  = "3"
	DUMMY4  = "4"
)

var mainConfig = map[string]map[string]string{
	LIBVIRT: {
		"inject_dir":    "libvirt/kvm/platform_config",
		"metadata_file": "libvirt/kvm/metadata/xml.tmplt",
		"topology_config_file":   "common/platform_config/topology_config.xml",
		"input_data_config_file": "libvirt/kvm/platform_config/input_data_config.xml",
	},
}

func Deploy(c *deployer.CommonData) error {
	if err := archutils.Extract(filepath.Join(c.RootDir, "comp/env.tgz"), c.RootDir); err != nil {
		return err
	}

	ui := c.Ui
	for {
		ui.SetLabel("Environment:")
		deployType := ui.Menu(4, "1", "Libvirt+Virsh",
			"2", "dummy2",
			"3", "dummy3",
			"4", "dummy4")

		switch deployType {
		case LIBVIRT:
			libvirtCreator := new(libvirt.FlowCreator)
			libvirtCreator.Filler = ImageFiller(c, mainConfig[LIBVIRT])
			libvirtCreator.SrcMetadataFile = filepath.Join(c.RootDir, "env", mainConfig[LIBVIRT]["metadata_file"])
			libvirtCreator.TopologyConfigFile = filepath.Join(c.RootDir, "env", mainConfig[LIBVIRT]["topology_config_file"])
			libvirtCreator.InputDataConfigFile = filepath.Join(c.RootDir, "env", mainConfig[LIBVIRT]["input_data_config_file"])
			return main.Deploy(c, libvirtCreator)
		default:
			return nil
		}
	}
	return nil
}
