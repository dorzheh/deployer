package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwfilter"
	"github.com/dorzheh/deployer/utils/hwinfo"
	sshconf "github.com/dorzheh/infra/comm/common"
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
		ui.SetSize(8, len(defaultName)+10)
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

	if _, err := exec.LookPath("sshfs"); err != nil {
		ui.ErrorOutput("sshfs utility is not installed", 8, 14)
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
			run := utils.RunFunc(cfg)
			// verifying that user is able execute a command by using sudo
			_, err := run("uname")
			errCh <- err
		}()

		sleep, _ := time.ParseDuration("1s")
		err := ui.Wait("Trying to establish SSH connection to remote host.\nPlease wait...", sleep, errCh)
		if err == nil {
			break
		}
		ui.Output(gui.Warning, "Unable to establish SSH connection.\nPress <OK> to proceed", 7, 2)
	}
	return cfg
}

func UiNetworks(ui *gui.DialogUi, data *xmlinput.XMLInputData, allowedNics hwinfo.NICList) (*deployer.OutputNetworkData, error) {
	netMetaData := new(deployer.OutputNetworkData)
	netMetaData.Networks = make([]*xmlinput.Network, 0)
	netMetaData.NICLists = make([]hwinfo.NICList, 0)

	for _, net := range data.Networks {
		var modes []xmlinput.ConnectionMode
		if net.UiModeBinding != nil {
			mode, err := uiNetworkPolicySelector(ui, net.Name, net.UiModeBinding)
			if err != nil {
				return nil, err
			}
			modes = append(modes, mode)
		} else {
			for _, mode := range net.Modes {
				modes = append(modes, mode.Type)
			}
		}

		retainedNics, modePassthrough, err := hwfilter.NicsByType(allowedNics, modes)
		if err != nil {
			return nil, err
		}

		var list hwinfo.NICList
		switch {
		case net.MaxIfaces > 1 && retainedNics.Length() > 1:
			list, err = uiSelectMultipleNics(ui, retainedNics, &allowedNics, modePassthrough, net)
			if err != nil {
				return nil, err
			}
		default:
			list, err = uiSelectSingleNic(ui, retainedNics, &allowedNics, modePassthrough, net.Name)
			if err != nil {
				return nil, err
			}

		}
		netMetaData.Networks = append(netMetaData.Networks, net)
		netMetaData.NICLists = append(netMetaData.NICLists, list)
	}

	return netMetaData, nil
}

func uiSelectMultipleNics(ui *gui.DialogUi, selectedList hwinfo.NICList, fullList *hwinfo.NICList, modePassthrough bool, network *xmlinput.Network) (hwinfo.NICList, error) {
	var temp []string
	keeper := make(map[string]*hwinfo.NIC)
	indexInt := 1
	for _, nic := range selectedList {
		indexStr := strconv.Itoa(indexInt)
		temp = append(temp, indexStr, fmt.Sprintf("%-14s %-10s", nic.Name, nic.Desc), "off")
		keeper[indexStr] = nic
		indexInt++
	}

	sliceLength := len(temp)
	if sliceLength == 0 {
		return nil, errors.New("Make sure that XML file representing xmlinput is configured properly")
	}

	var selected []string
	for {
		ui.SetSize(sliceLength+5, 95)
		ui.SetLabel(fmt.Sprintf("Select interfaces for network \"%s\"", network.Name))
		selected = ui.Checklist(sliceLength, temp[0:]...)
		// for some reason Checklist returns slice of length 1 even nothing was selected
		if selected[0] == "" {
			continue
		}
		if uint(len(selected)) > network.MaxIfaces {
			continue
		}
		break
	}

	finalList := hwinfo.NewNICList()
	for _, index := range selected {
		nic := keeper[index]
		finalList.Add(nic)
		if modePassthrough && nic.Type == hwinfo.NicTypePhys {
			i, err := fullList.SearchIndexByPCI(nic.PCIAddr)
			if err != nil {
				return nil, err
			}
			fullList.Remove(i)
		}
	}
	return finalList, nil
}

func uiSelectSingleNic(ui *gui.DialogUi, selectedList hwinfo.NICList, fullList *hwinfo.NICList, modePassthrough bool, network string) (hwinfo.NICList, error) {
	var temp []string
	keeper := make(map[string]*hwinfo.NIC)
	indexInt := 1
	for _, nic := range selectedList {
		indexStr := strconv.Itoa(indexInt)
		temp = append(temp, indexStr, fmt.Sprintf("%-14s %-10s", nic.Name, nic.Desc))
		keeper[indexStr] = nic
		indexInt++
	}

	sliceLength := len(temp)
	if sliceLength == 0 {
		return nil, errors.New("Make sure that XML file representing xmlinput is configured properly")
	}
	ui.SetSize(sliceLength+5, 95)
	ui.SetLabel(fmt.Sprintf("Select interface for network \"%s\"", network))
	ifaceNum := ui.Menu(sliceLength, temp[0:]...)
	nic := keeper[ifaceNum]
	if modePassthrough && nic.Type == hwinfo.NicTypePhys {
		i, err := fullList.SearchIndexByPCI(nic.PCIAddr)
		if err != nil {
			return nil, err
		}
		fullList.Remove(i)
	}

	finalList := hwinfo.NewNICList()
	finalList.Add(nic)
	return finalList, nil
}

func uiNetworkPolicySelector(ui *gui.DialogUi, network string, modes []*xmlinput.Appearance) (xmlinput.ConnectionMode, error) {
	sliceLength := len(modes)
	if sliceLength == 0 {
		return xmlinput.ConTypeError, errors.New("ui_mode_selection is not set")
	}

	var temp []string
	index := 1
	for _, mode := range modes {
		temp = append(temp, strconv.Itoa(index), mode.Appear)
		index++
	}

	ui.SetSize(sliceLength+8, 50)
	ui.SetLabel(fmt.Sprintf("Network interface type for network \"%s\"", network))
	ifaceNumInt, err := strconv.Atoi(ui.Menu(sliceLength, temp[0:]...))
	if err != nil {
		return xmlinput.ConTypeError, err
	}
	return modes[ifaceNumInt-1].Type, nil
}

func UiGatherHWInfo(ui *gui.DialogUi, hidriver deployer.HostinfoDriver, sleepInSec string, remote bool) error {
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- hidriver.Init()
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

func UiRAMSize(ui *gui.DialogUi, installedRamInMb, defaultRamInMb, reqMinimumRamInMb, reqMaximumRamInMb uint) uint {
	var amountUint uint
	maxIsSet := true
	if reqMaximumRamInMb == 0 {
		maxIsSet = false
		reqMaximumRamInMb = installedRamInMb
	}
	var defaultRamInMbStr string
	if defaultRamInMb != 0 {
		defaultRamInMbStr = strconv.Itoa(int(defaultRamInMb))
	}
	for {
		msg := fmt.Sprintf("Enter VM RAM size(minimum %dMB, maximum %dMB)", reqMinimumRamInMb, reqMaximumRamInMb)
		ui.SetSize(8, len(msg)+10)
		ui.SetLabel(msg)
		amountStr := ui.Inputbox(defaultRamInMbStr)
		amountInt, err := strconv.Atoi(amountStr)
		if err != nil {
			continue
		}
		amountUint = uint(amountInt)
		if amountUint > installedRamInMb {
			ui.Output(gui.Warning, "Required RAM exceeds host machine available memory.\nPress <OK> to proceed", 7, 2)
			continue
		}
		if amountUint < reqMinimumRamInMb {
			ui.Output(gui.Warning, fmt.Sprintf("Minimum RAM requirement is %dMB.\nPress <OK> to proceed", reqMinimumRamInMb), 7, 2)
			continue
		}
		if maxIsSet && amountUint > reqMaximumRamInMb {
			ui.Output(gui.Warning, fmt.Sprintf("Maximum RAM requirement is %dMB.\nPress <OK> to proceed", reqMaximumRamInMb), 7, 2)
			continue
		}
		break
	}
	return amountUint
}

func UiCPUs(ui *gui.DialogUi, installedCpus, defaultCpus, reqMinimumCpus, reqMaximumCpus uint) uint {
	var amountUint uint
	maxIsSet := true
	if reqMaximumCpus == 0 {
		maxIsSet = false
		reqMaximumCpus = installedCpus
	}
	var defaultCpusStr string
	if defaultCpus != 0 {
		defaultCpusStr = strconv.Itoa(int(defaultCpus))
	}
	for {
		msg := fmt.Sprintf("Enter VM number of vCPUs(minimum %d, maximum %d)", reqMinimumCpus, reqMaximumCpus)
		ui.SetSize(8, len(msg)+10)
		ui.SetLabel(msg)
		amountStr := ui.Inputbox(defaultCpusStr)
		amountInt, err := strconv.Atoi(amountStr)
		if err != nil {
			continue
		}

		amountUint = uint(amountInt)
		if amountUint < reqMinimumCpus {
			ui.Output(gui.Warning, fmt.Sprintf("Minimum vCPUs requirement is %d.\nPress <OK> to proceed", reqMinimumCpus), 7, 2)
			continue
		}
		if maxIsSet && amountUint > reqMaximumCpus {
			ui.Output(gui.Warning, fmt.Sprintf("Amount of vCPUs exceeds maximum supported vCPUs(%d).\nPress <OK> to proceed", reqMaximumCpus), 7, 2)
			continue
		}
		if amountUint > installedCpus {
			if !UiVCPUsOvercommit(ui, installedCpus) {
				continue
			}
		}
		break
	}
	return amountUint
}

func UiVCPUsOvercommit(ui *gui.DialogUi, installedCpus uint) bool {
	ui.SetSize(7, 85)
	ui.SetLabel(fmt.Sprintf("WARNING:The host only has %d CPUs.Overcommitting vCPUs can reduce performance!\nWould you like to proceed?", installedCpus))
	return ui.Yesno()
}
