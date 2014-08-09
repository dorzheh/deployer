package ui

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/deployer/utils"
	infrautils "github.com/dorzheh/infra/utils"
)

func UiValidateUser(ui *gui.DialogUi, userId int) {
	if err := infrautils.ValidateUserID(userId); err != nil {
		ui.ErrorOutput(err.Error(), 6, 14)
	}
}

func UiWelcomeMsg(ui *gui.DialogUi, name string) {
	msg := "Welcome to the " + name + " Deployment Procedure!"
	ui.SetSize(6, len(msg)+5)
	ui.Msgbox(msg)
}

func UiDeploymentResult(ui *gui.DialogUi, msg string, err error) {
	if err != nil {
		ui.ErrorOutput(err.Error(), 8, 14)
	}
	width := len(msg) + 2
	ui.Output(gui.Success, msg, 6, width)
}

func UiHostName(ui *gui.DialogUi) (hostname string) {
	for {
		ui.SetSize(8, 30)
		ui.SetLabel("Set hostname")
		hostname = ui.Inputbox("")
		if err := infrautils.SetHostname(hostname); err != nil {
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
		ui.SetLabel("Appliance name")
		name = ui.Inputbox(defaultName)
		if name != "" {
			name = strings.Replace(name, ".", "-", -1)
			if driver != nil {
				if driver.DomainExists(name) {
					ui.Output(gui.Warning, "domain "+name+" exists.\nPress <OK> and choose another name", 8, 10)
					continue
				}
			}
			break
		}
	}
	return name
}

func UiImagePath(ui *gui.DialogUi, defaultLocation string, remote bool) (location string) {
	for {
		if remote {
			location = ui.GetFromInput("Location to store the image on remote server", defaultLocation)
			break
		}
		location = ui.GetPathToDirFromInput(defaultLocation, "Select location to store the image")
		if _, err := os.Stat(location); err == nil {
			break
		}
	}
	return
}

func UiRemoteMode(ui *gui.DialogUi) bool {
	ui.SetLabel("Deployment Mode")
	answer := ui.Menu(2, "1", "Local", "2", "Remote")
	if answer == "1" {
		return false
	}
	return true
}

func UiRemoteParams(ui *gui.DialogUi) (ip string, port string, user string, passwd string, keyFile string) {
	ip = ui.GetIpFromInput("Remote server IP")
	port = "22"
	for {
		port = ui.GetFromInput("SSH port", port)
		if portDig, err := strconv.Atoi(port); err == nil {
			if portDig < 65536 {
				break
			}
		}
	}
	user = ui.GetFromInput("Username for logging into the host "+ip, "root")
	for {
		ui.SetLabel("Authentication method")
		switch ui.Menu(2, "1", "Password", "2", "Private key") {
		case "1":
			passwd = ui.GetPasswordFromInput(ip, user)
			return
		case "2":
			location := ui.GetPathToFileFromInput("Path to ssh private key file")
			if _, err := os.Stat(location); err == nil {
				return
			}
		}
	}
	return
}

func UiNetworks(ui *gui.DialogUi, info []*utils.NicInfo, networks ...string) (map[string]*utils.NicInfo, error) {
	newMap := make(map[string]*utils.NicInfo)
	for _, net := range networks {
		nic, err := uiGetNicInfo(ui, &info, net)
		if err != nil {
			return nil, err
		}
		newMap[net] = nic
	}
	nextIndex := len(networks)
	for {
		ui.SetSize(5, 60)
		if len(networks) == 0 {
			ui.SetLabel("Would you like to configure network?")
		} else {
			ui.SetLabel("Would you like to configure additional network?")
		}
		if !ui.Yesno() {
			break
		}
		net := fmt.Sprintf("#%d", nextIndex)
		nic, err := uiGetNicInfo(ui, &info, net)
		if err != nil {
			return nil, err
		}
		newMap[net] = nic
	}
	return newMap, nil
}

func uiGetNicInfo(ui *gui.DialogUi, info *[]*utils.NicInfo, network string) (*utils.NicInfo, error) {
	var temp []string
	index := 0
	for _, n := range *info {
		index += 1
		temp = append(temp, strconv.Itoa(index), fmt.Sprintf("%-14s %-10s", n.Name, n.Desc))
	}
	sliceLength := len(temp)
	var ifaceNumStr string
	var err error
	for {
		ui.SetSize(sliceLength+2, 95)
		ui.SetLabel(fmt.Sprintf("Select interface for network \"%s\"", network))
		ifaceNumStr = ui.Menu(sliceLength, temp[0:]...)
		if ifaceNumStr != "" {
			break
		}
	}
	ifaceNumInt, err := strconv.Atoi(ifaceNumStr)
	if err != nil {
		return nil, err
	}
	index = ifaceNumInt - 1
	nic := (*info)[index]
	if nic.Type == utils.NicTypePhys {
		tempInfo := *info
		tempInfo = append(tempInfo[:index], tempInfo[index+1:]...)
		*info = tempInfo
	}
	return nic, nil
}

func UiGatherHWInfo(ui *gui.DialogUi, hw *utils.HwInfoParser, sleepInSec string, remote bool) error {
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- hw.Parse()
	}()
	sleep, err := time.ParseDuration(sleepInSec)
	if err != nil {
		return err
	}

	var msg string
	if remote {
		msg = "Gathering harwdare information from remote host.\nPlease wait..."
	} else {
		msg = "Gathering hardware information.\nPlease wait..."
	}
	return ui.Wait(msg, sleep, errCh)
}

func UiConfirmation(ui *gui.DialogUi, buf *bytes.Buffer, height int) {
	buf.WriteString("\n\nPress <OK> to proceed or <CTRL+C> to exit")
	ui.SetSize(height, 100)
	ui.Msgbox(buf.String())
}
