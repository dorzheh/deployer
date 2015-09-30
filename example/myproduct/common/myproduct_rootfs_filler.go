//
// Implementing "image.Rootfs" interface
//
package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/builder/content"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
	archutils "github.com/dorzheh/infra/utils/archutils"
)

const (
	ERROR_SKIP_A = "create_inode: failed to change uid and gids on"
	ERROR_SKIP_B = "Cannot change ownership to uid 0, gid 0"
)

type rootfsFiller struct {
	pathToKitDir               string
	pathToRootfsSquashfs       string
	pathToKernelArchive        string
	pathToKernelModulesArchive string
	pathToApplArchive          string
	pathToConfigDir            string
	extractApplImage           bool
}

func (f *rootfsFiller) CustomizeRootfs(pathToRootfsMp string) error {
	unsquashfs, err := exec.LookPath("unsquashfs")
	if err != nil {
		unsquashfs = filepath.Join(f.pathToKitDir, "install/x86_64/bin/unsquashfs")
	}

	path := filepath.Join(pathToRootfsMp, "rootfs")
	exec.Command(unsquashfs, "-dest", path, f.pathToRootfsSquashfs).Run()
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return utils.FormatError(err)
	}
	for _, cont := range dir {
		exec.Command("cp", "-a", filepath.Join(path, cont.Name()), pathToRootfsMp).CombinedOutput()
	}
	if err := os.RemoveAll(path); err != nil {
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
			if !strings.Contains(err.Error(), ERROR_SKIP_B) {
				return utils.FormatError(err)
			}
		}

		modulesDir, err := filepath.Glob(filepath.Join(pathToRootfsMp, "lib/modules/*"))
		if err != nil {
			return utils.FormatError(err)
		}
		if len(modulesDir) == 0 {
			return errors.New("modules not found")
		}

		kernelVersion := filepath.Base(modulesDir[0])
		if err := os.Symlink("/boot/vmlinuz-"+kernelVersion, filepath.Join(pathToRootfsMp, "vmlinuz")); err != nil {
			return utils.FormatError(err)
		}
		if err := os.Symlink("/boot/initrd.img-"+kernelVersion, filepath.Join(pathToRootfsMp, "initrd.img")); err != nil {
			return utils.FormatError(err)
		}
	}

	pathToCommonDir := filepath.Join(f.pathToKitDir, "comp/env/common/config")
	fd, err := os.Stat(pathToCommonDir)
	if err == nil && fd.IsDir() {
		if err := content.Customize(pathToRootfsMp, pathToCommonDir); err != nil {
			return utils.FormatError(err)
		}
	}
	if f.pathToConfigDir != "" {
		if err := content.Customize(pathToRootfsMp, f.pathToConfigDir); err != nil {
			return utils.FormatError(err)
		}
	}
	return nil
}

// InstallApp is responsible for application installation
func (f *rootfsFiller) InstallApp(pathToRootfsMp string) error {
	if err := archutils.Extract(f.pathToApplArchive, filepath.Join(pathToRootfsMp, "mnt/cf")); err != nil {
		return utils.FormatError(err)
	}
	if f.extractApplImage {
		return extractApplImage(pathToRootfsMp)
	}
	return nil
}

// RunHooks is responsible for executing hooks before the image is being cleaned up
func (f *rootfsFiller) RunHooks(pathToRootfsMp string) error {
	return nil
}

// extractAppImage is responsible for extracting application image in a chroot environment
func extractApplImage(pathRootMp string) error {
	if err := os.Chdir(filepath.Join(pathRootMp, "mnt/cf")); err != nil {
		return utils.FormatError(err)
	}
	if err := exec.Command("/bin/bash", "-c", "./myapp*").Run(); err != nil {
		return fmt.Errorf("extracting application image : %v", err.Error())
	}
	return os.Chdir("/")
}

func ImageFiller(data *deployer.CommonData, configDir string) deployer.RootfsFiller {
	return &rootfsFiller{
		pathToKitDir:               data.RootDir,
		pathToRootfsSquashfs:       filepath.Join(data.RootDir, "comp/rootfs.squashfs"),
		pathToKernelArchive:        filepath.Join(data.RootDir, "comp/kernel.tgz"),
		pathToKernelModulesArchive: filepath.Join(data.RootDir, "comp/modules.tgz"),
		pathToApplArchive:          filepath.Join(data.RootDir, "comp/appl.tgz"),
		pathToConfigDir:            filepath.Join(data.RootDir, configDir),
		extractApplImage:           false,
	}
}
