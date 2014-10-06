package bundle

import (
	"fmt"
	"strconv"

	gui "github.com/dorzheh/deployer/ui/dialog_ui"
)

func uiBundleConfig(ui *gui.DialogUi, configs []*Config) (*Config, error) {
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
		configNumStr = ui.Menu(sliceLength+6, temp[0:]...)
		if configNumStr != "" {
			break
		}
	}

	configNumInt, err := strconv.Atoi(configNumStr)
	if err != nil {
		return nil, err
	}
	if configNumInt == advIndex {
		return nil, nil
	}
	return configs[configNumInt-1], nil
}
