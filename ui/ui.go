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
	sshconf "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/ssh"
	infrautils "github.com/dorzheh/infra/utils"
)

func UiValidateUser(ui *gui.DialogUi, userId int) {
	if err := infrautils.ValidateUserID(userId); err != nil {
		ui.ErrorOutput(err.Error(), 6, 14)
	}
}

func UiWelcomeMsg(ui *gui.DialogUi, name string) {
	msg := "Welcome to the " + name + " deployment procedure!"
	ui.SetSize(6, len(msg)+5)
	ui.Msgbox(msg)
}

func UiDeploymentResult(ui *gui.DialogUi, msg string, err error) {
	if err != nil {
		ui.ErrorOutput(err.Error(), 8, 14)
	}
	width := len(msg) + 5
	ui.Output(gui.None, msg, 6, width)
}

func UiApplianceName(ui *gui.DialogUi, defaultName string, driver deployer.Driver) string {
	var name string
	for {
		ui.SetSize(8, 30)
		ui.SetLabel("Virtual machine name")
		name = ui.Inputbox(defaultName)
		if name != "" {
			name = strings.Replace(name, ".", "-", -1)
			if driver != nil {
				if driver.DomainExists(name) {
					ui.Output(gui.Warning, "domain "+name+" already exists.\nPress <OK> and choose another name", 7, 2)
					continue
				}
			}
			break
		}
	}
	return name
}

func UiImagePath(ui *gui.DialogUi, defaultLocation string, remote bool) string {
	if remote {
		return ui.GetFromInput("Select directory on remote server to install the VA image on", defaultLocation)
	}
	var location string
	for {
		location = ui.GetPathToDirFromInput("Select directory to install the VA image on", defaultLocation)
		if _, err := os.Stat(location); err == nil {
			break
		}
	}
	return location
}

func UiRemoteMode(ui *gui.DialogUi) bool {
	ui.SetLabel("Deployment Mode")
	answer := ui.Menu(2, "1", "Local", "2", "Remote")
	if answer == "1" {
		return false
	}
	return true
}

func UiSshConfig(ui *gui.DialogUi) *sshconf.Config {
	errCh := make(chan error)
	defer close(errCh)
	cfg := new(sshconf.Config)

	for {
		cfg.Host = ui.GetIpFromInput("Remote server IP")
		cfg.Port = "22"
		for {
			cfg.Port = ui.GetFromInput("SSH port", cfg.Port)
			if portDig, err := strconv.Atoi(cfg.Port); err == nil {
				if portDig < 65536 {
					break
				}
			}
		}

		cfg.User = ui.GetFromInput("Username for logging into the host "+cfg.Host, "root")
		ui.SetLabel("Authentication method")
		switch ui.Menu(2, "1", "Password", "2", "Private key") {
		case "1":
			cfg.Password = ui.GetPasswordFromInput(cfg.Host, cfg.User, false)
		case "2":
			cfg.PrvtKeyFile = ui.GetPathToFileFromInput("Path to ssh private key file")

		}

		go func() {
			_, err := ssh.NewSshConn(cfg)
			errCh <- err
		}()

		sleep, _ := time.ParseDuration("1s")
		err := ui.Wait("Trying to establish SSH connection to remote host.\nPlease wait...", sleep, errCh)
		if err != nil {
			ui.Output(gui.Warning, "Unable to establish SSH connection.\nPress <OK> to proceed", 7, 2)
		} else {
			break
		}
	}
	return cfg
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
	ui.SetSize(5, 60)
	if len(networks) == 0 {
		ui.SetLabel("Would you like to configure network?")
	} else {
		ui.SetLabel("Would you like to configure additional network?")
	}
	if ui.Yesno() {
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
	ui.SetSize(sliceLength+5, 95)
	ui.SetLabel(fmt.Sprintf("Select interface for network \"%s\"", network))
	ifaceNumInt, err := strconv.Atoi(ui.Menu(sliceLength, temp[0:]...))
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
