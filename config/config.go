package config

import (
	"fmt"

	"github.com/dorzheh/deployer/builder/image"
)

func StorageConfig(pathToMainImage string, configIndex image.ConfigIndex, sconf *image.Storage, diskSizeMbSlice []int) (*image.Config, error) {
	conf, err := sconf.IndexToConfig(configIndex)
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
