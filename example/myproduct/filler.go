package myproduct

import (
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/deployer"
)

func ImageFiller(data *deployer.CommonData, mainConfig map[string]string) image.Rootfs {
	return &rootfsFiller{
		pathToKitDir:               data.RootDir,
		pathToRootfsArchive:        filepath.Join(data.RootDir, "comp/rootfs.tgz"),
		pathToKernelArchive:        filepath.Join(data.RootDir, "comp/kernel.tgz"),
		pathToKernelModulesArchive: filepath.Join(data.RootDir, "comp/modules.tgz"),
		pathToApplArchive:          filepath.Join(data.RootDir, "comp/appl.tgz"),
		pathToInjectDir:            filepath.Join(data.RootDir, "comp/env", mainConfig["inject_dir"]),
		extractApplImage:           false,
	}
}

func InstanceFiller(data *deployer.CommonData, mainConfig map[string]string) image.Rootfs {
	return &rootfsFiller{
		pathToKitDir:               data.RootDir,
		pathToApplArchive:          filepath.Join(data.RootDir, "comp/appl.tgz"),
		pathToInjectDir:            filepath.Join(data.RootDir, "comp/env", mainConfig["inject_dir"]),
		pathToRootfsArchive:        "",
		pathToKernelArchive:        "",
		pathToKernelModulesArchive: "",
		extractApplImage:           false,
	}
}
