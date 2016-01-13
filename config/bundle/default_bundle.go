package bundle

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/ui"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/deployer/utils"
)

// Configuration example:

// <?xml version="1.0" encoding="UTF-8"?>
// <bundle>
//   <config>
// 		<name>Test1</name>
// 		<cpus>2</cpus>
//      <ram>4096</ram>
//      <!-- storage configuration index -->
//      <storage_config_index>0</storage_config_index>
//   </config>
//   <config>
// 	     <name>Test2</name>
// 	     <cpus>2</cpus>
// 	     <ram>8192</ram>
// 	     <storage_config_index>1</storage_config_index>
//   </config>
//   <config>
// 	     <name>Test3</name>
// 	     <cpus>8</cpus>
//        <ram>16384</ram>
//        <storage_config_index>2</storage_config_index>
//   </config>
//   <advanced_config>true</advanced_config>
// </bundle>
//

type Config struct {
	Name               string            `xml:"name"`
	CPUs               int               `xml:"cpus"`
	RAM                int               `xml:"ram_mb"`
	StorageConfigIndex image.ConfigIndex `xml:"storage_config_index"`
}

type DefaultBundle struct {
	Configs        []*Config `xml:"config"`
	AdvancedConfig bool      `xml:"advanced_config"`
}

func (b *DefaultBundle) Parse(d *deployer.CommonData, hidriver deployer.HostinfoDriver, xid *xmlinput.XMLInputData) (map[string]interface{}, error) {
	hostRamsizeMb, err := hidriver.RAMSize()
	if err != nil {
		return nil, utils.FormatError(err)
	}

	configs := b.getConfigs(hostRamsizeMb)
	amountOfConfigs := len(configs)
	if amountOfConfigs == 0 {
		return nil, utils.FormatError(errors.New("no eligable configuration is available for the host"))
	}

	installedCpus, err := hidriver.CPUs()
	if err != nil {
		return nil, utils.FormatError(err)
	}
	for {
		c, err := uiBundleConfig(d.Ui, configs, b.AdvancedConfig)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		if c == nil {
			break
		}
		if c.CPUs > installedCpus {
			if !ui.UiVCPUsOvercommit(d.Ui, installedCpus) {
				continue
			}
		}
		m := make(map[string]interface{})
		m["name"] = c.Name
		m["cpus"] = c.CPUs
		m["ram_mb"] = c.RAM
		m["storage_config_index"] = c.StorageConfigIndex
		return m, nil
	}
	return nil, nil
}

func (b *DefaultBundle) getConfigs(ramsizeMb int) []*Config {
	configs := make([]*Config, 0)
	for _, c := range b.Configs {
		if c.RAM <= ramsizeMb {
			configs = append(configs, c)
		}
	}
	return configs
}

func uiBundleConfig(ui *gui.DialogUi, configs []*Config, advancedConfig bool) (*Config, error) {
	var temp []string
	index := 0
	for _, c := range configs {
		index += 1
		temp = append(temp, strconv.Itoa(index),
			fmt.Sprintf("%-15s [ vCPU %-2d | RAM %-3dMB]", c.Name, c.CPUs, c.RAM))
	}

	advIndex := index + 1
	temp = append(temp, strconv.Itoa(advIndex), "Custom configuration")

	sliceLength := len(temp)
	var configNumStr string
	var err error
	for {
		ui.SetSize(sliceLength+6, 50)
		ui.SetLabel("Select Virtual Machine configuration")
		configNumStr, err = ui.Menu(sliceLength+6, temp[0:]...)
		if err != nil {
			return nil, err
		}
		if configNumStr != "" {
			break
		}
	}

	configNumInt, err := strconv.Atoi(configNumStr)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if configNumInt == advIndex {
		return nil, nil
	}
	return configs[configNumInt-1], nil
}
