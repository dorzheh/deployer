package ui

import (
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

func UiApplianceName(ui *gui.DialogUi, defaultName string, driver deployer.EnvDriver) string {
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
		ui.Output(gui.Warning, "Unable to establish SSH connection.\nPress <OK> to proceed.", 7, 2)
	}
	return cfg
}

func UiNetworks(ui *gui.DialogUi, data *xmlinput.XMLInputData, allowedNics hwinfo.NICList) (*deployer.OutputNetworkData, error) {
	netMetaData := new(deployer.OutputNetworkData)
	netMetaData.Networks = make([]*xmlinput.Network, 0)
	netMetaData.NICLists = make([]hwinfo.NICList, 0)

	for i, net := range data.Networks.Configs {
		if !net.Mandatory {
			ui.SetSize(5, 60)
			if i == 0 {
				ui.SetLabel(fmt.Sprintf("Would you like to configure network (%s)?", net.Name))
			} else {
				ui.SetLabel("Would you like to configure additional network?")
			}
			if !ui.Yesno() {
				// do not configure rest of the networks
				break
			}
		}

		var modes []xmlinput.ConnectionMode
		if net.UiModeBinding == nil || len(net.UiModeBinding) == 0 {
			for _, mode := range net.Modes {
				modes = append(modes, mode.Type)
			}
		} else {
			var err error
			modes, err = uiNetworkPolicySelector(ui, net.Name, net.UiModeBinding)
			if err != nil {
				return nil, utils.FormatError(err)
			}
		}

		retainedNics, modePassthrough, err := hwfilter.NicsByType(allowedNics, modes)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		if len(retainedNics) == 0 {
			ui.Output(gui.Warning, "no interfaces have been found.\nPress <OK> to return to menu.", 7, 2)
			continue
		}

		var list hwinfo.NICList
		switch {
		case net.MaxIfaces > 1 && retainedNics.Length() > 1:
			list, err = uiSelectMultipleNics(ui, retainedNics, &allowedNics, modePassthrough, net)
			if err != nil {
				return nil, utils.FormatError(err)
			}
		default:
			list, err = uiSelectSingleNic(ui, retainedNics, &allowedNics, modePassthrough, net.Name)
			if err != nil {
				return nil, utils.FormatError(err)
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
	var selected []string
	for {
		ui.SetSize(sliceLength+5, 95)
		ui.SetLabel(fmt.Sprintf("Select interfaces for network \"%s\"", network.Name))
		selected = ui.Checklist(sliceLength, temp[0:]...)
		// for some reason Checklist returns slice of length 1 even nothing was selected
		if selected[0] == "" {
			continue
		}
		if len(selected) > network.MaxIfaces {
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
				return nil, utils.FormatError(err)
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
	ui.SetSize(sliceLength+5, 95)
	ui.SetLabel(fmt.Sprintf("Select interface for network \"%s\"", network))
	ifaceNum := ui.Menu(sliceLength, temp[0:]...)
	nic := keeper[ifaceNum]
	if modePassthrough && nic.Type == hwinfo.NicTypePhys {
		i, err := fullList.SearchIndexByPCI(nic.PCIAddr)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		fullList.Remove(i)
	}

	finalList := hwinfo.NewNICList()
	finalList.Add(nic)
	return finalList, nil
}

func uiNetworkPolicySelector(ui *gui.DialogUi, network string, modes []*xmlinput.Appearance) ([]xmlinput.ConnectionMode, error) {
	matrix := make(map[string][]xmlinput.ConnectionMode)
	for _, mode := range modes {
		if _, ok := matrix[mode.Appear]; !ok {
			matrix[mode.Appear] = make([]xmlinput.ConnectionMode, 0)
		}
		matrix[mode.Appear] = append(matrix[mode.Appear], mode.Type)
	}

	var temp []string
	index := 1
	for appear, _ := range matrix {
		temp = append(temp, strconv.Itoa(index), appear)
		index++
	}

	length := len(matrix)
	ui.SetSize(length+8, 50)
	ui.SetLabel(fmt.Sprintf("Network interface type for network \"%s\"", network))
	resultInt, err := strconv.Atoi(ui.Menu(length, temp[0:]...))
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if resultInt != 1 {
		resultInt++
	}
	return matrix[temp[resultInt]], nil
}

func UiGatherHWInfo(ui *gui.DialogUi, hidriver deployer.HostinfoDriver, sleepInSec string, remote bool) error {
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- hidriver.Init()
	}()
	sleep, err := time.ParseDuration(sleepInSec)
	if err != nil {
		return utils.FormatError(err)
	}

	var msg string
	if remote {
		msg = "Gathering harwdare information from remote host.\nPlease wait..."
	} else {
		msg = "Gathering hardware information.\nPlease wait..."
	}
	return ui.Wait(msg, sleep, errCh)
}

type VmConfig struct {
	CPUs    int
	RamMb   int
	DisksMb []int
}

func UiVmConfig(ui *gui.DialogUi, driver deployer.HostinfoDriver, xidata *xmlinput.XMLInputData) (*VmConfig, error) {
	var installedRamMb int
	var err error
	conf := new(VmConfig)
	list := make([]string, 0)
	index := 1

	if xidata.CPU.Configure {
		cpuStr := fmt.Sprintf("  %-9s |  %d-%d", "CPU", xidata.CPU.Min, xidata.CPU.Max)
		list = []string{cpuStr, strconv.Itoa(index), "1", strconv.Itoa(xidata.CPU.Default), "1", "30", "6", "0", "0"}
		index++
	} else if xidata.CPU.Default > 0 {
		conf.CPUs = xidata.CPU.Default
	}
	if xidata.RAM.Configure {
		installedRamMb, err = driver.RAMSize()
		if err != nil {
			return nil, utils.FormatError(err)
		}
		if xidata.RAM.Max > installedRamMb || xidata.RAM.Max == xmlinput.UnlimitedAlloc {
			xidata.RAM.Max = installedRamMb
		}

		ramStr := fmt.Sprintf("  %-9s |  %d-%dG", "RAM", xidata.RAM.Min/1024, xidata.RAM.Max/1024)
		list = append(list, []string{ramStr, strconv.Itoa(index), "1", strconv.Itoa(xidata.RAM.Default / 1024), "2", "30", "6", "0", "0"}...)
		index++
	} else if xidata.RAM.Default > 0 {
		conf.RamMb = xidata.RAM.Default
	}
	if xidata.Disks.Configure {
		diskName := "Disk"
		for i, disk := range xidata.Disks.Configs {
			if i > 0 {
				diskName = strconv.Itoa(i) + "_" + strconv.Itoa(i)
			}
			diskStr := fmt.Sprintf("  %-9s |  %d-%dG", diskName, disk.Min/1024, disk.Max/1024)
			indexStr := strconv.Itoa(index)
			list = append(list, []string{diskStr, indexStr, "1", strconv.Itoa(disk.Default / 1024), indexStr, "30", "6", "0", "0"}...)
			index++
		}
	}
	str := "Virtual Machine Configuration\n------------------------------------------\n|	 Resource	 |	 Supported	 |	 Allocated	 |"

	installedCpus, err := driver.CPUs()
	if err != nil {
		return nil, utils.FormatError(err)
	}

	index--

	if index > 1 {

	MainLoop:
		for {
			ui.SetSize(13, 46)
			resultIndex := 0
			result := ui.Mixedform(str, list[0:]...)
			if len(result) < index {
				continue
			}
			if xidata.CPU.Configure {
				selectedCpus, err := strconv.Atoi(result[resultIndex])
				if err != nil {
					continue
				}
				if uiCpuNotOK(ui, selectedCpus, installedCpus, xidata.CPU.Min, xidata.CPU.Max) {
					continue
				}
				conf.CPUs = selectedCpus
				resultIndex++
			}
			if xidata.RAM.Configure {
				selectedRamGb, err := strconv.Atoi(result[resultIndex])
				if err != nil {
					continue
				}
				if uiRamNotOK(ui, selectedRamGb*1024, installedRamMb, xidata.RAM.Min, xidata.RAM.Max) {
					continue
				}
				conf.RamMb = selectedRamGb * 1024
				resultIndex++
			}
			if xidata.Disks.Configure {
				for _, disk := range xidata.Disks.Configs {
					selectedDiskSizeGb, err := strconv.Atoi(result[resultIndex])
					if err != nil {
						continue MainLoop
					}
					if uiDiskNotOK(ui, selectedDiskSizeGb*1024, disk.Min, disk.Max) {
						continue MainLoop
					}
					conf.DisksMb = append(conf.DisksMb, selectedDiskSizeGb*1024)
				}
			}
			break
		}
	}
	return conf, nil
}

func UiVCPUsOvercommit(ui *gui.DialogUi, installedCpus int) bool {
	ui.SetSize(7, 85)
	ui.SetLabel(fmt.Sprintf("WARNING:The host only has %d CPUs.Overcommitting vCPUs can reduce performance!\nWould you like to proceed?", installedCpus))
	return ui.Yesno()
}

func uiCpuNotOK(ui *gui.DialogUi, selectedCpus, installedCpus, minCpus, maxCpus int) bool {
	if selectedCpus < minCpus {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum vCPUs requirement is %d.\nPress <OK> to return to menu.", minCpus), 7, 2)
		return true
	}
	if selectedCpus > maxCpus {
		ui.Output(gui.Warning, fmt.Sprintf("Amount of vCPUs exceeds maximum supported vCPUs(%d).\nPress <OK> to return to menu.", maxCpus), 7, 2)
		return true
	}
	if selectedCpus > installedCpus {
		if !UiVCPUsOvercommit(ui, installedCpus) {
			return true
		}
	}
	return false
}

func uiRamNotOK(ui *gui.DialogUi, selectedRamInMb, installedRamMb, minRamInMb, maxRamInMb int) bool {
	if selectedRamInMb > installedRamMb {
		ui.Output(gui.Warning, "Required RAM exceeds host machine available memory.\nPress <OK> to return to menu.", 7, 2)
		return true
	}
	if selectedRamInMb < minRamInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum RAM requirement is %dGB.\nPress <OK> to return to menu.", minRamInMb/1024), 7, 2)
		return true
	}
	if selectedRamInMb > maxRamInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Maximum RAM requirement is %dGB.\nPress <OK> to return to menu.", maxRamInMb/1024), 7, 2)
		return true
	}
	return false
}

func uiDiskNotOK(ui *gui.DialogUi, selectedDiskInMb, minDiskInMb, maxDiskInMb int) bool {
	if selectedDiskInMb < minDiskInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum disk size requirement is %dGB.\nPress <OK> to return to menu.", minDiskInMb/1024), 7, 2)
		return true
	}
	if selectedDiskInMb > maxDiskInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Maximum disk size requirement is %dGB.\nPress <OK> to return to menu.", maxDiskInMb/1024), 7, 2)
		return true
	}
	return false
}
