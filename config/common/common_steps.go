// Intended for creating a common configuration
// required by any target
package common

import (
	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/controller"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
)

func RegisterSteps(d *deployer.CommonData, storageConfigFile string) (*deployer.CommonConfig, error) {
	var err error

	c := new(deployer.CommonConfig)
	c.StorageConfig, err = image.ParseConfigFile(storageConfigFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	controller.RegisterSteps(func() func() error {
		return func() error {
			var err error
			c.RemoteMode, err = gui.UiRemoteMode(d.Ui)
			return err
		}
	}())

	controller.RegisterSteps(func() func() error {
		return func() error {
			var err error
			if c.RemoteMode {
				c.SshConfig, err = gui.UiSshConfig(d.Ui)
				return err
			}
			return controller.SkipStep
		}
	}())

	controller.RegisterSteps(func() func() error {
		return func() error {
			var err error
			c.ExportDir, err = gui.UiImagePath(d.Ui, d.DefaultExportDir, c.RemoteMode)
			return err
		}
	}())
	return c, nil
}
