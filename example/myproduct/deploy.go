package myproduct

import (
	"path/filepath"

	main "github.com/dorzheh/deployer"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct/libvirt"
	"github.com/dorzheh/infra/utils"
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
	},
}

func Deploy(c *deployer.CommonData) error {
	if err := utils.Extract(filepath.Join(c.RootDir, "comp/env.tgz"), c.RootDir); err != nil {
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
			libvirtCreator.SrcMetadata = filepath.Join(c.RootDir, "env", mainConfig[LIBVIRT]["metadata_file"])
			return main.Deploy(c, libvirtCreator)
		default:
			return nil
		}
	}
	return nil
}
