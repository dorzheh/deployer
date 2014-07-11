package myproduct

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
)

func ImageFiller(data *deployer.CommonData, mainConfig map[string]string) image.Rootfs {
	return &RootfsFiller{
		PathToKitDir:               data.RootDir,
		PathToRootfsArchive:        filepath.Join(data.RootDir, "comp/rootfs.tgz"),
		PathToKernelArchive:        filepath.Join(data.RootDir, "comp/kernel.tgz"),
		PathToKernelModulesArchive: filepath.Join(data.RootDir, "comp/modules.tgz"),
		PathToApplArchive:          filepath.Join(data.RootDir, "comp/appl.tgz"),
		PathToInjectDir:            filepath.Join(data.RootDir, "comp/env", mainConfig["inject_dir"]),
		ExtractApplImage:           false,
	}
}

func InstanceFiller(data *deployer.CommonData, mainConfig map[string]string) image.Rootfs {
	return &RootfsFiller{
		PathToKitDir:               data.RootDir,
		PathToApplArchive:          filepath.Join(data.RootDir, "comp/appl.tgz"),
		PathToInjectDir:            filepath.Join(data.RootDir, "comp/env", mainConfig["inject_dir"]),
		PathToRootfsArchive:        "",
		PathToKernelArchive:        "",
		PathToKernelModulesArchive: "",
		ExtractApplImage:           false,
	}
}
