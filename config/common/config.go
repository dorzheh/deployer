// Intended for creating a common configuration
// required by either target
package common

import (
	"fmt"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
)

func CreateConfig(d *deployer.CommonData) *deployer.CommonConfig {
	c := new(deployer.CommonConfig)
	c.Ctrl = deployer.NewUiStepsController()
	c.Ctrl.RegisterSteps(func(*deployer.CommonConfig) func() error {
		return func() error {
			var err error
			c.RemoteMode, err = gui.UiRemoteMode(d.Ui)
			return err
		}
	}(c))

	c.Ctrl.RegisterSteps(func(*deployer.CommonConfig) func() error {
		return func() error {
			var err error
			if c.RemoteMode {
				c.SshConfig, err = gui.UiSshConfig(d.Ui)
				return err
			}
			return deployer.SkipStep
		}
	}(c))

	c.Ctrl.RegisterSteps(func(*deployer.CommonConfig) func() error {
		return func() error {
			var err error
			c.ExportDir, err = gui.UiImagePath(d.Ui, d.DefaultExportDir, c.RemoteMode)
			return err
		}
	}(c))

	return c
}

func StorageConfig(storageConfigFile, pathToMainImage string, configIndex image.ConfigIndex, diskSizeMbSlice []int) (*image.Config, error) {
	f, err := image.ParseConfigFile(storageConfigFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	conf, err := f.IndexToConfig(configIndex)
	if err != nil {
		return nil, err
	}

	deviceIndex := 0
	amountOfDisks := len(conf.Disks)
	for ; amountOfDisks != 0; amountOfDisks-- {
		if deviceIndex == 0 {
			conf.Disks[deviceIndex].Path = fmt.Sprintf("%s.%s", pathToMainImage, conf.Disks[deviceIndex].Type)
		} else {
			conf.Disks[deviceIndex].Path = fmt.Sprintf("%s_%d.%s", pathToMainImage, deviceIndex, conf.Disks[deviceIndex].Type)
		}
		if len(diskSizeMbSlice) > 0 {
			conf.Disks[deviceIndex].SizeMb = diskSizeMbSlice[deviceIndex]
		}
		deviceIndex++
	}
	return conf, nil
}
