//
// Implementing "image.Rootfs" interface
//
package myproduct

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/infra/utils/archutils"
)

type rootfsFiller struct {
	pathToKitDir               string
	pathToRootfsArchive        string
	pathToKernelArchive        string
	pathToKernelModulesArchive string
	pathToApplArchive          string
	pathToInjectDir            string
	extractApplImage           bool
}

func (f *rootfsFiller) MakeRootfs(pathToRootfsMp string) error {
	if err := archutils.Extract(f.pathToRootfsArchive, pathToRootfsMp); err != nil {
		return utils.FormatError(err)
	}
	if f.pathToKernelModulesArchive != "" && f.pathToKernelArchive != "" {
		errCh := make(chan error, 2)
		defer close(errCh)

		go func() {
			errCh <- archutils.Extract(f.pathToKernelArchive, filepath.Join(pathToRootfsMp, "boot"))
		}()
		go func() {
			errCh <- archutils.Extract(f.pathToKernelModulesArchive, filepath.Join(pathToRootfsMp, "lib/modules"))
		}()
		if err := utils.WaitForResult(errCh, 2); err != nil {
			return utils.FormatError(err)
		}

		modulesDir, err := filepath.Glob(filepath.Join(pathToRootfsMp, "lib/modules/*"))
		if err != nil {
			return utils.FormatError(err)
		}
		if len(modulesDir) == 0 {
			return utils.FormatError(errors.New("modules not found"))
		}

		kernelVersion := filepath.Base(modulesDir[0])
		if err := os.Symlink("/boot/vmlinuz-"+kernelVersion, filepath.Join(pathToRootfsMp, "vmlinuz")); err != nil {
			return utils.FormatError(err)
		}
		if err := os.Symlink("/boot/initrd.img-"+kernelVersion, filepath.Join(pathToRootfsMp, "initrd.img")); err != nil {
			return utils.FormatError(err)
		}
	}

	pathToEnvDir := filepath.Join(f.pathToKitDir, "env")
	pathToCommonDir := filepath.Join(pathToEnvDir, "common")
	fd, err := os.Stat(pathToCommonDir)
	if err == nil && fd.IsDir() {
		if err := image.Customize(pathToRootfsMp, pathToCommonDir); err != nil {
			return utils.FormatError(err)
		}
	}
	if f.pathToInjectDir != "" {
		if err := image.Customize(pathToRootfsMp, f.pathToInjectDir); err != nil {
			return utils.FormatError(err)
		}
	}
	return nil
}

// InstallApp is responsible for application installation
func (f *rootfsFiller) InstallApp(pathToRootfsMp string) error {
	if err := archutils.Extract(f.pathToApplArchive, filepath.Join(pathToRootfsMp, "mnt/compact_flash")); err != nil {
		return utils.FormatError(err)
	}
	if f.extractApplImage {
		return extractApplImage(pathToRootfsMp)
	}
	return nil
}

// extractApplImage is responsible for extracting application image
func extractApplImage(pathRootMp string) error {
	if err := os.Chdir(filepath.Join(pathRootMp, "/mnt/compact_flash")); err != nil {
		return utils.FormatError(err)
	}
	if err := exec.Command("/bin/bash", "-c", "./Myappl*").Run(); err != nil {
		return utils.FormatError(fmt.Errorf("extracting application image : %v", err.Error()))
	}
	return os.Chdir("/")
}
