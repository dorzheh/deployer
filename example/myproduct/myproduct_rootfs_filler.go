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

type RootfsFiller struct {
	PathToKitDir               string
	PathToRootfsArchive        string
	PathToKernelArchive        string
	PathToKernelModulesArchive string
	PathToApplArchive          string
	PathToInjectDir            string
	ExtractApplImage           bool
}

func (f *RootfsFiller) MakeRootfs(pathToRootfsMp string) error {
	if f.PathToRootfsArchive == "" {
		if f.PathToInjectDir != "" {
			if err := utils.RunPrePostScripts(filepath.Join(f.PathToInjectDir,
				"pre-deploy-scripts"), utils.PRE_SCRIPTS); err != nil {
				return err
			}
		}
	} else {
		if err := archutils.Extract(f.PathToRootfsArchive, pathToRootfsMp); err != nil {
			return err
		}
	}
	if f.PathToKernelModulesArchive != "" && f.PathToKernelArchive != "" {
		errCh := make(chan error, 2)
		defer close(errCh)

		go func() {
			errCh <- archutils.Extract(f.PathToKernelArchive, filepath.Join(pathToRootfsMp, "boot"))
		}()
		go func() {
			errCh <- archutils.Extract(f.PathToKernelModulesArchive, filepath.Join(pathToRootfsMp, "lib/modules"))
		}()
		if err := utils.WaitForResult(errCh, 2); err != nil {
			return err
		}

		modulesDir, err := filepath.Glob(filepath.Join(pathToRootfsMp, "lib/modules/*"))
		if err != nil {
			return err
		}
		if len(modulesDir) == 0 {
			return errors.New("modules not found")
		}

		kernelVersion := filepath.Base(modulesDir[0])
		if err := os.Symlink("/boot/vmlinuz-"+kernelVersion, filepath.Join(pathToRootfsMp, "vmlinuz")); err != nil {
			return err
		}
		if err := os.Symlink("/boot/initrd.img-"+kernelVersion, filepath.Join(pathToRootfsMp, "initrd.img")); err != nil {
			return err
		}
	}

	pathToEnvDir := filepath.Join(f.PathToKitDir, "env")
	pathToCommonDir := filepath.Join(pathToEnvDir, "common")
	fd, err := os.Stat(pathToCommonDir)
	if err == nil && fd.IsDir() {
		if err := image.Customize(pathToRootfsMp, pathToCommonDir); err != nil {
			return err
		}
	}
	if f.PathToInjectDir != "" {
		if err := image.Customize(pathToRootfsMp, f.PathToInjectDir); err != nil {
			return err
		}
	}
	return nil
}

// InstallApp is responsible for application installation
func (f *RootfsFiller) InstallApp(pathToRootfsMp string) error {
	if err := archutils.Extract(f.PathToApplArchive, filepath.Join(pathToRootfsMp, "mnt/cf")); err != nil {
		return err
	}
	if f.ExtractApplImage {
		return extractApplImage(pathToRootfsMp)
	}
	return nil
}

// extractApplImage is responsible for extracting application image
func extractApplImage(pathRootMp string) error {
	if err := os.Chdir(filepath.Join(pathRootMp, "/mnt/cf")); err != nil {
		return err
	}
	if err := exec.Command("/bin/bash", "-c", "./Myappl*").Run(); err != nil {
		return fmt.Errorf("extracting application image : %v", err.Error())
	}
	return os.Chdir("/")
}
