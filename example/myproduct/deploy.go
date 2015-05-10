package myproduct

import (
	"os"
	"path/filepath"

	main "github.com/dorzheh/deployer"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct/environments/libvirt/kvm"
	"github.com/dorzheh/infra/utils/archutils"
)

const (
	LIBVIRT_KVM = "1"
	DUMMY2      = "2"
	DUMMY3      = "3"
	DUMMY4      = "4"
)

var mainConfig = map[string]map[string]string{
	LIBVIRT_KVM: {
		"inject_dir":             "libvirt/kvm/config",
		"metadata_file":          "libvirt/kvm/metadata/xml.tmplt",
		"storage_config_file":    "common/config/storage_config.xml",
		"input_data_config_file": "libvirt/kvm/config/input_data_config.xml",
	},
}

// MainMenu is a strting point for dialog session
// Select appropriate target and run deployer
func Deploy(c *deployer.CommonData, prodType string) error {
	if err := archutils.Extract(filepath.Join(c.RootDir, "comp/env.tgz"), filepath.Join(c.RootDir, "comp")); err != nil {
		return err
	}

	ui := c.Ui
	for {
		ui.SetLabel("Environment:")
		deployType, _ := ui.Menu(4, "1", "Libvirt+Virsh(KVM)",
			"2", "dummy2",
			"3", "dummy3",
			"4", "dummy4")

		switch deployType {
		case LIBVIRT_KVM:
			libvirtCreator := new(kvm.FlowCreator)
			libvirtCreator.Filler = ImageFiller(c, mainConfig[prodType])
			libvirtCreator.SrcMetadataFile = filepath.Join(c.RootDir, "comp/env", mainConfig[prodType]["metadata_file"])
			libvirtCreator.BundleConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig[prodType]["bundle_config_file"])
			libvirtCreator.StorageConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig[prodType]["storage_config_file"])
			libvirtCreator.InputDataConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig[prodType]["input_data_config_file"])
			return main.Deploy(c, libvirtCreator)
		default:
			os.Exit(0)
		}
	}
	return nil
}
