package ui

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dorzheh/deployer/config"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/infra/utils"
)

func UiHostName(ui *gui.DialogUi) (hostname string) {
	for {
		ui.SetSize(8, 30)
		ui.SetLabel("New hostname: ")
		hostname = ui.Inputbox("")
		if err := utils.SetHostname(hostname); err != nil {
			ui.Output(gui.Warning, err.Error()+".Press <OK> to proceed", 8, 12)
			continue
		}
		break
	}
	return
}

func UiApplianceName(ui *gui.DialogUi, defaultName string, driver deployer.Driver) string {
	var name string
	for {
		ui.SetSize(8, 30)
		ui.SetLabel("Appliance name: ")
		name := ui.Inputbox(defaultName)
		if name != "" {
			name = strings.Replace(name, ".", "-", -1)
			if driver.DomainExists(name) {
				ui.Output(gui.Warning, "domain "+name+" exists.Press <OK> and choose another name", 8, 12)
				continue
			}
			break
		}
	}
	return name
}

// setImageLocation method sets location for the VA image
func UiImagePath(ui *gui.DialogUi, defaultLocation string) (location string) {
	for {
		ui.SetSize(6, 64)
		ui.Msgbox("Choose VA image location.\n\tPress <Ok> to proceed")
		location = ui.Dselect(defaultLocation)
		if _, err := os.Stat(location); err != nil {
			continue
		}
		break
	}
	return
}

func UiRemoteMode(ui *gui.DialogUi) bool {
	ui.SetLabel("Deployment Mode:")
	if answer := ui.Menu(2, "1", "Local", "2", "Remote"); answer == "1" {
		return false
	}
	return true
}

func UiRemoteParams(ui *gui.DialogUi) (string, string, string, string) {
	ip := ui.GetIpFromInput("Remote server IP:")
	user := ui.GetFromInput(ip+" superuser:", "root")
	var passwd string
	var keyFile string
	answer := ui.Menu(2, "1", "Password", "2", "Private key")
	if answer == "1" {
		passwd = ui.GetPasswordFromInput(ip, user)
	} else {
		for {
			ui.SetSize(6, 64)
			ui.Msgbox("Path to ssh private key.\n\tPress <Ok> to proceed")
			keyFile = ui.Fselect("")
			if _, err := os.Stat(keyFile); err != nil {
				continue
			}
			break
		}
	}
	return ip, user, passwd, keyFile
}

func UiNetworks(ui *gui.DialogUi, networks []string, info map[int]*config.NicInfo) (map[string]*config.NicInfo, error) {
	newMap := make(map[string]*config.NicInfo)
	var temp []string
	for _, n := range info {
		temp = append(temp, fmt.Sprintf("%s (driver type %s, %s)", n.Name, n.Driver, n.Desc))
	}

	sliceLength := len(temp)
	for _, net := range networks {
		var ifaceNumStr string
		for {
			ui.SetSize(10+sliceLength, 55)
			ui.SetLabel(fmt.Sprintf("Choose appropriate interface for \"%s\" network:", net))
			ifaceNumStr = ui.Menu(sliceLength, temp[0:]...)
			if ifaceNumStr != "" {
				break
			}
		}
		ifaceNumInt, err := strconv.Atoi(ifaceNumStr)
		if err != nil {
			return nil, err
		}
		newMap[net] = info[ifaceNumInt]
	}
	return newMap, nil
}

func UiConfirmation(ui *gui.DialogUi, buf *bytes.Buffer, height int) {
	buf.WriteString("\n\nPress <OK> to proceed or <CTRL+C> to exit")
	ui.SetSize(height, 100)
	ui.Msgbox(buf.String())
}
