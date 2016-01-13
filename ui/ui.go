package ui

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	main "github.com/dorzheh/deployer"
	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config"
	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/host_hwfilter"
	"github.com/dorzheh/deployer/utils/hwinfo/guest"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
	sshconf "github.com/dorzheh/infra/comm/common"
	infrautils "github.com/dorzheh/infra/utils"
)

// UiValidateUser intended for validate the ID
// of the user executing deployer binary
func UiValidateUser(ui *gui.DialogUi, userId int) {
	if err := infrautils.ValidateUserID(userId); err != nil {
		ui.Output(gui.Error, err.Error())
	}
}

// UiWelcomeMsg prints out appropriate welcome message
func UiWelcomeMsg(ui *gui.DialogUi, name string) {
	msg := "Welcome to the " + name + " deployment procedure!"
	ui.SetSize(6, len(msg)+5)
	ui.Msgbox(msg)
}

// UiEulaMsg prints out appropriate EULA message
func UiEulaMsg(ui *gui.DialogUi, pathToEula string) {
	ui.SetOkLabel("Agree")
	ui.SetExtraLabel("Disagree")
	ui.SetTitle("End User License Agreement")
	ui.SetSize(30, 80)
	if err := ui.Textbox(pathToEula); err != nil {
		os.Exit(1)
	}
}

func UiSelectEnv(c *deployer.CommonData, envList []string, envs []deployer.FlowCreator) error {
	var menuList []string
	indexInt := 1

	for _, env := range envList {
		menuList = append(menuList, strconv.Itoa(indexInt), env)
		indexInt++
	}

	envsNum := len(envs)
	dType := 0
	if envsNum > 1 {
		c.Ui.SetTitle("Select environment")
		c.Ui.SetSize(envsNum+7, 30)
		resStr, err := c.Ui.Menu(envsNum, menuList[0:]...)
		if err != nil && err.Error() == gui.DialogExit {
			os.Exit(0)
		}
		dType, err := strconv.Atoi(resStr)
		if err != nil {
			return utils.FormatError(err)
		}
		dType--
	}
	return main.Deploy(c, envs[dType])
}

func UiDeploymentResult(ui *gui.DialogUi, msg string, err error) {
	if err != nil {
		ui.Output(gui.Error, err.Error())
	}
	ui.Output(gui.Success, msg)
}

func UiApplianceName(ui *gui.DialogUi, defaultName string, driver deployer.EnvDriver) (string, error) {
	var name string
	var err error

	for {
		ui.SetSize(8, len(defaultName)+10)
		ui.SetTitle("Virtual machine name")
		ui.HelpButton(true)
		ui.SetHelpLabel("Back")
		name, err = ui.Inputbox(defaultName)
		if err != nil {
			return "", err
		}
		if name != "" {
			name = strings.Replace(name, ".", "-", -1)
			if driver != nil {
				if driver.DomainExists(name) {
					ui.Output(gui.Warning, "domain "+name+" already exists.", "Press <OK> to return to menu.")
					continue
				}
			}
			break
		}
	}
	return name, nil
}

func UiImagePath(ui *gui.DialogUi, defaultLocation string, remote bool) (string, error) {
	if remote {
		return ui.GetFromInput("Select directory on remote server to install the VA image on", defaultLocation, "Back", "")
	}

	var location string
	var err error

	for {
		location, err = ui.GetPathToDirFromInput("Select directory to install the VA image on", defaultLocation, "Back", "")
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(location); err == nil {
			break
		}
	}
	return location, nil
}

func UiRemoteMode(ui *gui.DialogUi) (bool, error) {
	ui.SetTitle("Deployment Mode")
	ui.SetSize(9, 28)
	answer, err := ui.Menu(2, "1", "Local", "2", "Remote")
	if err != nil {
		return false, err
	}
	if answer == "1" {
		return false, nil
	}

	if _, err := exec.LookPath("sshfs"); err != nil {
		ui.Output(gui.Error, "sshfs utility is not installed")
	}
	return true, nil
}

func UiSshConfig(ui *gui.DialogUi) (*sshconf.Config, error) {
	cfg := new(sshconf.Config)
	cfg.Port = "22"
	cfg.User = "root"
	origlist := []string{"IP      : ", "1", "1", "", "1", "10", "22", "0", "0",
		"SSH Port: ", "2", "1", cfg.Port, "2", "10", "22", "0", "0",
		"Username: ", "3", "1", cfg.User, "3", "10", "22", "0", "0"}

MainLoop:
	for {
		ui.HelpButton(true)
		ui.SetHelpLabel("Back")
		reslist, err := ui.Mixedform("Remote session configuration", false, origlist[0:]...)
		if err != nil {
			return nil, err
		}
		if len(reslist) < 3 {
			continue
		}
		if net.ParseIP(reslist[0]) == nil {
			continue
		}
		cfg.Host = reslist[0]

		portDig, err := strconv.Atoi(reslist[1])
		if err != nil {
			return nil, utils.FormatError(err)
		}
		if portDig > 65535 {
			continue
		}

	AuthLoop:
		for {
			ui.SetTitle("Authentication method")
			ui.SetSize(9, 18)
			ui.HelpButton(true)
			ui.SetHelpLabel("Back")
			val, err := ui.Menu(2, "1", "Password", "2", "Private key")
			if err != nil {
				switch err.Error() {
				case gui.DialogMoveBack:
					continue MainLoop
				case gui.DialogExit:
					os.Exit(1)
				}
			}

			switch val {
			case "1":
				cfg.Password, err = ui.GetPasswordFromInput(cfg.Host, cfg.User, "Back", "", false)
			case "2":
				cfg.PrvtKeyFile, err = ui.GetPathToFileFromInput("Path to ssh private key file", "Back", "")
			}
			if err != nil {
				switch err.Error() {
				case gui.DialogMoveBack:
					continue AuthLoop
				case gui.DialogExit:
					os.Exit(1)
				}
			}
			break MainLoop
		}
	}

	run := utils.RunFunc(cfg)
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		// verifying that user is able execute a command by using sudo
		_, err := run("uname")
		errCh <- err
	}()

	if err := ui.Wait("Trying to establish SSH connection to remote host.\nPlease wait...", time.Second*1, time.Second*5, errCh); err != nil {
		ui.Output(gui.Warning, "Unable to establish SSH connection.", "Press <OK> to return to menu.")
		goto MainLoop
	}
	return cfg, nil
}

func UiNetworks(ui *gui.DialogUi, data *xmlinput.XMLInputData, allowedNics host.NICList, gconf *guest.Config) error {
	guestPciSlotCounter := data.GuestNic.PCI.FirstSlot
	lastGuestPciSlotCounter := guestPciSlotCounter
	portCounter := 1
	lastPortCounter := portCounter
	i := 0

MainLoop:
	for i < len(data.Networks.Configs) {
		net := data.Networks.Configs[i]
	PolicyLoop:
		for {
			var modes []xmlinput.ConnectionMode

			if net.UiModeBinding == nil || len(net.UiModeBinding) == 0 {
				for _, mode := range net.Modes {
					modes = append(modes, mode.Type)
				}
			} else {
				var err error
				modes, err = uiNetworkPolicySelector(ui, net)
				if err != nil {
					switch err.Error() {
					case gui.DialogMoveBack:
						gconf.Networks = gconf.Networks[:i]
						gconf.NICLists = gconf.NICLists[:i]
						portCounter = lastPortCounter - 1
						guestPciSlotCounter = lastGuestPciSlotCounter - 1
						if i == 0 {
							return err
						}
						i--
						continue MainLoop
					case gui.DialogNext:
						i++
						continue MainLoop
					case gui.DialogExit:
						os.Exit(1)
					default:
						return err
					}

				}
			}

			retainedNics, err := host_hwfilter.NicsByType(allowedNics, modes)
			if err != nil {
				return utils.FormatError(err)
			}
			if len(retainedNics) == 0 {
				ui.Output(gui.Warning, "No interfaces have been found.", "Press <OK> to return to menu.")
				continue MainLoop
			}
			if net.UiResetCounter {
				portCounter = 1
			}
			list, err := uiNicSelectMenu(ui, data, &portCounter, &guestPciSlotCounter, retainedNics, net, i)
			if err != nil {
				switch err.Error() {
				case gui.DialogMoveBack:
					if i == 0 {
						return err
					}
					gconf.Networks = gconf.Networks[:i]
					gconf.NICLists = gconf.NICLists[:i]
					continue PolicyLoop
				case gui.DialogExit:
					os.Exit(1)
				}
			}

			gconf.Networks = append(gconf.Networks, net)
			gconf.NICLists = append(gconf.NICLists, list)
			lastPortCounter = portCounter
			lastGuestPciSlotCounter = guestPciSlotCounter
			i++
			break
		}
	}
	return nil
}

func uiNicSelectMenu(ui *gui.DialogUi, data *xmlinput.XMLInputData, guestPortCounter *int,
	guestPciSlotCounter *int, hnics host.NICList, net *xmlinput.Network, index int) (guest.NICList, error) {
	list := make([]string, 0)
	keeper := make(map[string]*host.NIC)
	indexStrToInt := make(map[string][]int)
	indexInt := 1
	for _, hnic := range hnics {
		indexStr := strconv.Itoa(indexInt)
		list = append(list, indexStr, fmt.Sprintf("%-22s%-15s%-68s%-10s", hnic.Name, hnic.PCIAddr, hnic.Desc, uiNUMAIntToString(hnic.NUMANode)))
		keeper[indexStr] = hnic
		if indexInt == 1 {
			// index 0 - element index in the list
			// index 1 - element counter
			// index 2 - PCI slot number represented as integer
			indexStrToInt[indexStr] = []int{1, 0, 0}
		} else {
			indexStrToInt[indexStr] = []int{indexInt*2 - 1, 0, 0}
		}
		indexInt++
	}

	listLength := len(list)
	gnics := guest.NewNICList()
	var disjuncNicVendor string
	var disjuncNicModel string

	for {
		if index > 0 {
			ui.HelpButton(true)
			ui.SetHelpLabel("Back")
		}
		width := uiHeaderSelectNics(ui)
		ui.SetSize(listLength+8, width+5)
		ui.SetTitle(fmt.Sprintf("Select interface for network \"%s\"", net.Name))
		nicNumStr, err := ui.Menu(listLength+5, list[0:]...)
		if err != nil {
			if err.Error() == gui.DialogNext {
				if len(gnics) == 0 && net.Optional == false {
					continue
				}
				break
			}
			*guestPciSlotCounter -= len(gnics)
			*guestPortCounter -= len(gnics)
			return nil, err
		}

		hnic := keeper[nicNumStr]
		// verify that we need to proceed with "disjunction"
		if net.NicsDisjunction && (hnic.Type == host.NicTypePhys || hnic.Type == host.NicTypePhysVF) &&
			(hnic.Vendor != disjuncNicVendor && hnic.Model != disjuncNicModel) {
			// back to menu in case "disjunction" entries already exist
			if host_hwfilter.NicDisjunctionFound(hnic, data.HostNics.Allowed) && disjuncNicVendor != "" {
				msg := fmt.Sprintf("'%s' cannot be selected alongside '%s %s'", hnic.Desc, disjuncNicVendor, disjuncNicModel)
				ui.Output(gui.Warning, msg, "Press <OK> to return to menu.")
				continue
			}
			// set the new entries
			disjuncNicVendor = hnic.Vendor
			disjuncNicModel = hnic.Model
		}

		delNicCounter := indexStrToInt[nicNumStr][1]
		// counter for the NIC found,we should remove the NIC object and its references
		if delNicCounter > 0 {
			// find the guest NIC object in the guest NICs list
			if _, index, err := gnics.NicByHostNicObj(hnic); err == nil {
				// if found :
				// - remove the object
				// - decrement guestPortCounter
				// - decriment guestPciSlotCounter
				gnics.Remove(index)
				if *guestPortCounter > 0 {
					*guestPortCounter--
				}
				if *guestPciSlotCounter > data.GuestNic.PCI.FirstSlot {
					*guestPciSlotCounter--
				}
				// update the list with the new entry containing deselected NIC
				list[indexStrToInt[nicNumStr][0]] = fmt.Sprintf("%-22s%-15s%-68s%-10s", hnic.Name, hnic.PCIAddr, hnic.Desc, uiNUMAIntToString(hnic.NUMANode))
				// - reset element counter
				// - reset PCI slot number
				indexStrToInt[nicNumStr][1] = 0
				indexStrToInt[nicNumStr][2] = 0
			}
			// iterate over the map and update entries
			for nicIndex, data := range indexStrToInt {
				nicCounter := data[1]
				if nicCounter > delNicCounter {
					tmpNic := keeper[nicIndex]
					nicCounter--
					list[data[0]] = fmt.Sprintf("%-22s%-15s%-68s%-9s%d", tmpNic.Name, tmpNic.PCIAddr,
						tmpNic.Desc, uiNUMAIntToString(tmpNic.NUMANode), nicCounter)
					indexStrToInt[nicIndex][1] = nicCounter
					indexStrToInt[nicIndex][2]--
					if gnic, _, err := gnics.NicByHostNicObj(tmpNic); err == nil {
						gnic.PCIAddr.Slot = utils.IntToHexString(indexStrToInt[nicIndex][2])
					}
				}
			}
			// clear "disjunction" entries in case gnics slice is empty
			if gnics.Length() == 0 {
				disjuncNicVendor = ""
				disjuncNicModel = ""
			}
			continue
		}
		// create new guest NIC, set his  number and counter
		indexStrToInt[nicNumStr][1] = *guestPortCounter
		indexStrToInt[nicNumStr][2] = *guestPciSlotCounter
		list[indexStrToInt[nicNumStr][0]] = fmt.Sprintf("%-22s%-15s%-68s%-9s%d", hnic.Name, hnic.PCIAddr,
			hnic.Desc, uiNUMAIntToString(hnic.NUMANode), *guestPortCounter)
		gnic := guest.NewNIC()
		gnic.Network = net.Name
		gnic.PCIAddr.Domain = data.PCI.Domain
		gnic.PCIAddr.Bus = data.PCI.Bus
		gnic.PCIAddr.Slot = utils.IntToHexString(*guestPciSlotCounter)
		gnic.PCIAddr.Function = data.PCI.Function
		gnic.HostNIC = hnic
		gnics.Add(gnic)
		*guestPciSlotCounter++
		*guestPortCounter++
	}
	return gnics, nil
}

func uiNUMAIntToString(numaInt int) string {
	var numaStr string
	if numaInt == -1 {
		numaStr = "N/A"
	} else {
		numaStr = strconv.Itoa(numaInt)
	}
	return numaStr
}

func uiHeaderSelectNics(ui *gui.DialogUi) int {
	str := " ___________________________________________________________________________________________________________________________"
	width := len(str)
	str += "\n|____________________________________________________HOST__________________________________________________________|___VM___|"

	str += fmt.Sprintf("\n|________%s________|__%s__|_______________ %s _________________|__%s__|__%s__|", "Port Name", "PCI Address", "Network Interface Description", "NUMA", "Port")
	ui.SetLabel(str)
	ui.SetExtraLabel("Next")
	ui.SetOkLabel("Select")
	ui.HelpButton(true)
	ui.SetHelpLabel("Back")
	return width
}

func uiNetworkPolicySelector(ui *gui.DialogUi, net *xmlinput.Network) ([]xmlinput.ConnectionMode, error) {
	matrix := make(map[string][]xmlinput.ConnectionMode)
	for _, mode := range net.UiModeBinding {
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
	ui.SetTitle(fmt.Sprintf("Network interface type for network \"%s\"", net.Name))
	ui.HelpButton(true)
	ui.SetHelpLabel("Back")
	if net.Optional {
		ui.SetExtraLabel("Skip")
	}
	val, err := ui.Menu(length, temp[0:]...)
	if err != nil {
		return nil, err
	}
	resultInt, err := strconv.Atoi(val)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if resultInt != 1 {
		resultInt++
	}
	return matrix[temp[resultInt]], nil
}

func UiGatherHWInfo(ui *gui.DialogUi, hidriver deployer.HostinfoDriver, remote bool) error {
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- hidriver.Init()
	}()

	var msg string
	if remote {
		msg = "Gathering harwdare information from remote host.\nPlease wait..."
	} else {
		msg = "Gathering hardware information from local host.\nPlease wait..."
	}
	return ui.Wait(msg, time.Second*2, 0, errCh)
}

func UiVmConfig(ui *gui.DialogUi, driver deployer.HostinfoDriver, xidata *xmlinput.XMLInputData,
	pathToMainImage string, sconf *image.Storage, conf *guest.Config) error {
	var installedRamMb int
	var maxRAM int
	var err error

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
			return utils.FormatError(err)
		}
		if xidata.RAM.Max > installedRamMb || xidata.RAM.Max == xmlinput.UnlimitedAlloc {
			maxRAM = installedRamMb
		} else {
			maxRAM = xidata.RAM.Max
		}

		ramStr := fmt.Sprintf("  %-9s |  %d-%dG", "RAM", xidata.RAM.Min/1024, maxRAM/1024)
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
	str := " ______________________________________\n|	 Resource	 |		Maximum		|		Allocated		|"
	installedCpus, err := driver.CPUs()
	if err != nil {
		return utils.FormatError(err)
	}

	index--

	if index > 1 {

	MainLoop:
		for {
			ui.SetSize(11, 46)
			ui.SetTitle("Virtual Machine configuration")
			ui.HelpButton(true)
			ui.SetHelpLabel("Back")
			resultIndex := 0
			result, err := ui.Mixedform(str, false, list[0:]...)
			if err != nil {
				return err
			}
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
				selectedRamMb, err := utils.FloatStringToInt(result[resultIndex], 1024)
				if err != nil {
					continue MainLoop
				}
				if uiRamNotOK(ui, selectedRamMb, installedRamMb, xidata.RAM.Min, maxRAM) {
					continue
				}
				conf.RamMb = selectedRamMb
				resultIndex++
			}
			if xidata.Disks.Configure {
				disks := make([]int, 0)
				for _, disk := range xidata.Disks.Configs {
					selectedDiskSizeMb, err := utils.FloatStringToInt(result[resultIndex], 1024)
					if err != nil {
						continue MainLoop
					}
					if uiDiskNotOK(ui, selectedDiskSizeMb, disk.Min, disk.Max) {
						continue MainLoop
					}
					disks = append(disks, selectedDiskSizeMb*1024)
					resultIndex++
				}
				if conf.Storage, err = config.StorageConfig(pathToMainImage, 0, sconf, disks); err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func UiVCPUsOvercommit(ui *gui.DialogUi, installedCpus int) bool {
	ui.SetSize(8, 75)
	ui.SetTitle(gui.Warning)
	ui.SetLabel(fmt.Sprintf("\nThe host only has %d CPUs.Overcommitting vCPUs can reduce performance!\nWould you like to proceed?", installedCpus))
	return ui.Yesno()
}

func uiCpuNotOK(ui *gui.DialogUi, selectedCpus, installedCpus, minCpus, maxCpus int) bool {
	if selectedCpus < minCpus {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum vCPUs requirement is %d.", minCpus), "Press <OK> to return to menu.")
		return true
	}
	if selectedCpus > maxCpus {
		ui.Output(gui.Warning, fmt.Sprintf("Amount of vCPUs exceeds maximum supported vCPUs(%d).", maxCpus), "Press <OK> to return to menu.")
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
		ui.Output(gui.Warning, "Required RAM exceeds host machine available memory.", "Press <OK> to return to menu.")
		return true
	}
	if selectedRamInMb < minRamInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum RAM requirement is %dGB.", minRamInMb/1024), "Press <OK> to return to menu.")
		return true
	}
	if selectedRamInMb > maxRamInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Maximum RAM requirement is %dGB.", maxRamInMb/1024), "Press <OK> to return to menu.")
		return true
	}
	return false
}

func uiDiskNotOK(ui *gui.DialogUi, selectedDiskInMb, minDiskInMb, maxDiskInMb int) bool {
	if selectedDiskInMb < minDiskInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Minimum disk size requirement is %dGB.", minDiskInMb/1024), "Press <OK> to return to menu.")
		return true
	}
	if selectedDiskInMb > maxDiskInMb {
		ui.Output(gui.Warning, fmt.Sprintf("Maximum disk size requirement is %dGB.", maxDiskInMb/1024), "Press <OK> to return to menu.")
		return true
	}
	return false
}

func uiNUMATopologyHeader(ui *gui.DialogUi, c *guest.Config) string {
	ui.HelpButton(true)
	ui.SetHelpLabel("Back")
	ui.SetTitle("VA NUMA Configuration")
	ui.SetExtraLabel("Edit")

	var hdr string
	for _, n := range c.NUMAs {
		for _, nic := range n.NICs {
			if nic.HostNIC.Type == host.NicTypePhys || nic.HostNIC.Type == host.NicTypePhysVF {
				hdr += fmt.Sprintf("\nNUMA %d: %s", nic.HostNIC.NUMANode, nic.HostNIC.PCIAddr)
			}
		}
	}

	hdr += "\n"
	if hdr != "\n" {
		hdr = " \n---------------- PCI Devices Topology ---------------" + hdr
		hdr += "-----------------------------------------------------\n\n"
	}

	hdr += " __________________ CPU/NUMA Topology ________________\n"
	hdr += "|____________ VA ___________|_________ Host __________|\n"
	hdr += "|____ vCPU ___|___ vNUMA ___|_________CPU(s) _________|"
	return hdr
}

func uiShowNumaTopologyHelpMsg(ui *gui.DialogUi) {
	msg := "CPU Pinning Help\n"
	msg += "----------------\n\n"
	msg += "Example 1 : one to one pinning\n\n"
	msg += " _____________________________________\n"
	msg += "| VA vNUMA ID : 0_____________________|\n"
	msg += "| Host CPU(s) : 0_____________________|\n"
	msg += "|_____________________________________|\n\n"
	msg += " __________________ CPU/NUMA Topology ________________\n"
	msg += "|____________ VA ___________|_________ Host __________|\n"
	msg += "|____ vCPU ___|___ vNUMA ___|_________CPU(s) _________|\n"
	msg += "|______ 0 ____|_____ 0 _____|__________ 0 ____________|\n\n"
	msg += "VA CPU 0 will be pinned to the host CPU 0\n\n\n"
	msg += "Example 2 : pinning to a range of the host CPUs\n\n"
	msg += " _____________________________________\n"
	msg += "| VA vNUMA ID : 0_____________________|\n"
	msg += "| Host CPU(s) : 0-3___________________|\n"
	msg += "|_____________________________________|\n\n"
	msg += " __________________ CPU/NUMA Topology ________________\n"
	msg += "|____________ VA ___________|_________ Host __________|\n"
	msg += "|____ vCPU ___|___ vNUMA ___|_________CPU(s) _________|\n"
	msg += "|______ 0 ____|_____ 0 _____|__________ 0,1,2,3_______|\n\n"
	msg += "VA CPUs will be pinned to the host CPUs 0,1,2 and 3\n\n\n"
	msg += "Example 3 : pinning to a list of the host CPUs\n\n"
	msg += " _____________________________________\n"
	msg += "| VA vNUMA ID : 0_____________________|\n"
	msg += "| Host CPU(s) : 0,1,2,3,8_____________|\n"
	msg += "|_____________________________________|\n\n"
	msg += " __________________ CPU/NUMA Topology ________________\n"
	msg += "|____________ VA ___________|_________ Host __________|\n"
	msg += "|____ vCPU ___|___ vNUMA ___|_________CPU(s) _________|\n"
	msg += "|______ 0 ____|_____ 0 _____|_______ 0,1,2,3,8 _______|\n\n"
	msg += "VA CPUs will be pinned to the host CPUs 0,1,2,3 and 8\n\n"
	ui.Msgbox(msg)
}

func UiWarningOnOptimizationFailure(ui *gui.DialogUi, warningStr string) bool {
	ui.SetTitle(gui.Warning)
	ui.SetSize(10, 80)
	ui.SetLabel("Virtual machine configuration can not be optimized.\n" + warningStr + "\n\nDo you want to continue?")
	return ui.Yesno()
}

func UiNUMATopology(ui *gui.DialogUi, c *guest.Config, d deployer.EnvDriver, totalCpusOnHost int) error {
	var list []string

MainLoop:
	for {
		list = make([]string, 0)
		tempData := make(map[int]map[int]string)

		for _, n := range c.NUMAs {
			keys := make([]int, 0)
			for vcpu, _ := range n.CPUPin {
				keys = append(keys, vcpu)
			}

			sort.Ints(keys)
			var hostCpu string
			for _, k := range keys {
				if len(n.CPUPin[k]) > 1 {
					if len(n.CPUPin[k]) == totalCpusOnHost {
						hostCpu = "0-" + strconv.Itoa(totalCpusOnHost-1)
					} else {
						var tmpStrSlice []string
						for _, c := range n.CPUPin[k] {
							tmpStrSlice = append(tmpStrSlice, strconv.Itoa(c))
						}
						hostCpu = strings.Join(tmpStrSlice, ",")
					}
				} else {
					hostCpu = strconv.Itoa(n.CPUPin[k][0])
				}

				//keyStr := strconv.Itoa(k)
				tempData[k] = make(map[int]string)
				tempData[k][n.CellID] = hostCpu
			}
		}

		// we need to represent sorted vCPU IDs and not vNUMA IDs
		keys := make([]int, 0)
		for k, _ := range tempData {
			keys = append(keys, k)
		}

		sort.Ints(keys)
		for _, key := range keys {
			for k, v := range tempData[key] {
				list = append(list, strconv.Itoa(key), fmt.Sprintf("%-10s%-18d%-7s", " ", k, v))
			}
		}

		ui.SetLabel(uiNUMATopologyHeader(ui, c))
		result, err := ui.Menu(len(list), list[0:]...)
		if err == nil {
			break
		}
		if err.Error() != gui.DialogNext {
			return err
		}

	InternalLoop:
		for {
			ui.SetTitle("VA vCPU Configuration")
			ui.SetExtraLabel("Help")

			resultInt, err := strconv.Atoi(result)
			if err != nil {
				return err
			}

			var vnumaStr string
			var cpusStr string
			for k, v := range tempData[resultInt] {
				vnumaStr = strconv.Itoa(k)
				cpusStr = v
			}

			resultInt--

			var vnumaPredecStr string
			for k, _ := range tempData[resultInt] {
				vnumaPredecStr = strconv.Itoa(k)
			}

			var lst []string
			label := "Set affinity for vCPU " + result
			if vnumaStr == vnumaPredecStr && d.Id() == "QEMU-Libvirt" {
				lst = []string{"VA vNUMA ID : ", "1", "1", vnumaStr, "1", "15", "2", "0", "2"}
				label += "\n\nIMPORTANT! Some QEMU versions do not support\n" +
					"disjoint NUMA CPU ranges therefore vNUMA configuration\n" +
					"is disabled for this vCPU.\n"
			} else {
				lst = []string{"VA vNUMA ID : ", "1", "1", vnumaStr, "1", "15", "2", "0", "0"}
			}

			lst = append(lst, "Host CPU(s) : ", "2", "1", cpusStr, "2", "15", "30", "0", "0")
			r, err := ui.Mixedform(label, false, lst[0:]...)
			if err != nil {
				if err.Error() == gui.DialogNext {
					uiShowNumaTopologyHelpMsg(ui)
					continue
				}
				return err
			}
			if len(r) < 2 {
				continue
			}

			vcpuInt, err := strconv.Atoi(result)
			if err != nil {
				continue
			}

			vnumaInt, err := strconv.Atoi(r[0])
			if err != nil {
				ui.Output(gui.Warning, "Illegal input \""+r[0]+"\"", "Press <OK> to return to menu.")
				continue
			}
			if err := verifyRange(vnumaInt, len(c.NUMAs)); err != nil {
				ui.Output(gui.Warning, err.Error(), "Press <OK> to return to menu.")
				continue
			}

			hostCpus := r[1]
			cpus := make([]int, 0)
			if strings.Contains(hostCpus, ",") {
				for _, e := range strings.Split(hostCpus, ",") {
					if strings.Contains(e, "-") {
						cpus, err = splitByHypen(e, totalCpusOnHost)
						if err != nil {
							ui.Output(gui.Warning, err.Error(), "Press <OK> to return to menu.")
							continue InternalLoop
						}
					} else {
						cpu, err := strconv.Atoi(e)
						if err != nil {
							ui.Output(gui.Warning, "Illegal input \""+e+"\"", "Press <OK> to return to menu.")
							continue InternalLoop
						}
						if err := verifyRange(cpu, totalCpusOnHost); err != nil {
							ui.Output(gui.Warning, err.Error(), "Press <OK> to return to menu.")
							continue InternalLoop
						}
						cpus = append(cpus, cpu)
					}
				}
			} else if strings.Contains(hostCpus, "-") {
				cpus, err = splitByHypen(hostCpus, totalCpusOnHost)
				if err != nil {
					ui.Output(gui.Warning, err.Error(), "Press <OK> to return to menu.")
					continue
				}
			} else {
				cpu, err := strconv.Atoi(hostCpus)
				if err != nil {
					ui.Output(gui.Warning, "Illegal input \""+hostCpus+"\"", "Press <OK> to return to menu.")
					continue
				}
				if err := verifyRange(cpu, totalCpusOnHost); err != nil {
					ui.Output(gui.Warning, err.Error(), "Press <OK> to return to menu.")
					continue
				}
				cpus = append(cpus, cpu)
			}

			// delete the old entry
			for _, n := range c.NUMAs {
				for vcpu, _ := range n.CPUPin {
					if vcpu == vcpuInt {
						delete(n.CPUPin, vcpu)
					}
				}
			}

			sort.Ints(cpus)
			// set the new entry
			c.NUMAs[vnumaInt].CPUPin[vcpuInt] = cpus
			continue MainLoop
		}
		break
	}
	return nil
}

func splitByHypen(e string, totalCpusOnHost int) ([]int, error) {
	cpus := make([]int, 0)

	firstlast := strings.Split(e, "-")
	firstInt, err := strconv.Atoi(firstlast[0])
	if err != nil {
		return cpus, errors.New("Illegal CPU number.")
	}
	if err := verifyRange(firstInt, totalCpusOnHost); err != nil {
		return cpus, err
	}

	lastInt, err := strconv.Atoi(firstlast[1])
	if err != nil {
		return cpus, errors.New("Illegal CPU number.")
	}
	if err := verifyRange(lastInt, totalCpusOnHost); err != nil {
		return cpus, err
	}

	lastInt++
	for i := firstInt; i < lastInt; i++ {
		cpus = append(cpus, i)
	}
	return cpus, nil
}

func verifyRange(number, maxNumber int) error {
	if number < 0 || number > maxNumber-1 {
		return fmt.Errorf("The value (%d) is out of range", number)
	}
	return nil
}
