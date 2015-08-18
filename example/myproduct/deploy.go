package myproduct

import (
	"path/filepath"

	"github.com/dorzheh/deployer/deployer"
	libvirt_kvm "github.com/dorzheh/deployer/example/myproduct/env/libvirt/kvm"
	"github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/infra/utils/archutils"
)

var mainConfig = map[string]map[string]string{
	"libvirt_kvm": {
		"config_dir":             "libvirt/kvm/config",
		"metadata_dir":           "libvirt/kvm/metadata",
		"metadata_file":          "libvirt/kvm/metadata/domain.tmplt.xml",
		"storage_config_file":    "common/config/storage_config.xml",
		"input_data_config_file": "libvirt/kvm/config/input_data_config.xml",
	},
}

// MainMenu is a strting point for dialog session
// Select appropriate target and run deployer
func Deploy(c *deployer.CommonData) error {
	if err := archutils.Extract(filepath.Join(c.RootDir, "comp/env.tgz"), filepath.Join(c.RootDir, "comp")); err != nil {
		return err
	}

	libvirtCreator := new(libvirt_kvm.FlowCreator)
	libvirtCreator.Filler = ImageFiller(c, mainConfig["libvirt_kvm"])
	libvirtCreator.SrcMetadataFile = filepath.Join(c.RootDir, "comp/env", mainConfig["libvirt_kvm"]["metadata_file"])
	libvirtCreator.BundleConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig["libvirt_kvm"]["bundle_config_file"])
	libvirtCreator.StorageConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig["libvirt_kvm"]["storage_config_file"])
	libvirtCreator.InputDataConfigFile = filepath.Join(c.RootDir, "comp/env", mainConfig["libvirt_kvm"]["input_data_config_file"])
	libvirtCreator.MetadataDir = filepath.Join(c.RootDir, "comp/env", mainConfig["libvirt_kvm"]["metadata_dir"])

	return ui.UiDeploy(c, []string{"Libvirt(KVM)"}, []deployer.FlowCreator{libvirtCreator})
}
