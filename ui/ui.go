package ui

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/deployer/utils"
	infra "github.com/dorzheh/infra/utils"
)

func UiNewSession() *gui.DialogUi {
	return gui.NewDialogUi()
}

func UiValidateUser(ui *gui.DialogUi, userId int) {
	if err := infra.ValidateUserID(userId); err != nil {
		ui.ErrorOutput(err.Error(), 6, 14)
	}
}

func UiWelcomeMsg(ui *gui.DialogUi, name string) {
	ui.SetSize(6, 49)
	ui.Msgbox("Welcome to the " + name + " Deployment Procedure!")
}

func UiDeploymentResult(ui *gui.DialogUi, err error) {
	if err != nil {
		ui.ErrorOutput(err.Error(), 8, 14)
	}
	ui.Output(gui.Success, "deployment process completed.", 6, 14)
}

func UiHostName(ui *gui.DialogUi) (hostname string) {
	for {
		ui.SetSize(8, 30)
		ui.SetLabel("New hostname: ")
		hostname = ui.Inputbox("")
		if err := infra.SetHostname(hostname); err != nil {
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
		name = ui.Inputbox(defaultName)
		if name != "" {
			name = strings.Replace(name, ".", "-", -1)
			if driver != nil {
				if driver.DomainExists(name) {
					ui.Output(gui.Warning, "domain "+name+" exists.Press <OK> and choose another name", 8, 12)
					continue
				}
			}
			break
		}
	}
	return name
}

func UiImagePath(ui *gui.DialogUi, defaultLocation string) (location string) {
	for {
		ui.SetSize(6, 64)
		ui.Msgbox("The next step allows to choose location for the image.\n\tPress <Ok> to proceed")
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

func UiRemoteParams(ui *gui.DialogUi) (string, string, string, string, string) {
	ip := ui.GetIpFromInput("Remote server IP:")
	port := "22"
	for {
		port = ui.GetFromInput("SSH port:", port)
		if portDig, err := strconv.Atoi(port); err == nil {
			if portDig < 65536 {
				break
			}
		}
	}
	user := ui.GetFromInput(ip+" user:", "root")
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
	return ip, port, user, passwd, keyFile
}

func UiNetworks(ui *gui.DialogUi, info map[int]*utils.NicInfo, networks ...string) (map[string]*utils.NicInfo, error) {
	newMap := make(map[string]*utils.NicInfo)
	var temp []string
	index := 0
	for _, n := range info {
		index += 1
		temp = append(temp, strconv.Itoa(index), fmt.Sprintf("%-14s %-10s", n.Name, n.Desc))
	}

	sliceLength := len(temp)
	var ifaceNumStr string
	for _, net := range networks {
		var ifaceNumStr string
		for {
			ui.SetSize(sliceLength+2, 95)
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
		newMap[net] = info[ifaceNumInt-1]
	}

	nextIndex := len(networks) + 1
	for {
		ui.SetSize(5, 60)
		ui.SetLabel("Would you like to configure additional network?")
		if ui.Yesno() {
			for {
				ui.SetSize(sliceLength+2, 95)
				ui.SetLabel(fmt.Sprintf("Choose appropriate interface for the network #%d:", nextIndex))
				ifaceNumStr = ui.Menu(sliceLength, temp[0:]...)
				if ifaceNumStr != "" {
					break
				}
			}
			ifaceNumInt, err := strconv.Atoi(ifaceNumStr)
			if err != nil {
				return nil, err
			}
			newMap[ifaceNumStr] = info[ifaceNumInt-1]
			nextIndex++
		} else {
			break
		}
	}
	return newMap, nil
}

func UiConfirmation(ui *gui.DialogUi, buf *bytes.Buffer, height int) {
	buf.WriteString("\n\nPress <OK> to proceed or <CTRL+C> to exit")
	ui.SetSize(height, 100)
	ui.Msgbox(buf.String())
}
