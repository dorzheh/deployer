package ui

import (
	//"bytes"
	//"errors"
	//"fmt"
	//"io/ioutil"
	"os"
	//"os/exec"
	//"path/filepath"
	//"strconv"
	//"strings"

	"github.com/dorzheh/infra/utils"
	gui "github.com/dorzheh/ui/dialog-ui"
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

func UiRemote(ui *gui.DialogUi) (string, string, string, string) {
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

// setBridges binds bridge against appropriate network
// Returns error/nil
//func (vac *vaConfig) setBridges() error {
//	var installedBridges []string
//	var err error
//	if vac.vaInstallType == LOCAL {
//		installedBridges, err = utils.GetLocalBridges()
//		if err != nil {
//			return err
//		}
//	} else {
//		installedBridges, err = utils.GetRemoteBridges(vac.rsConfig.remoteIpAddress,
//			vac.rsConfig.remoteUser, vac.rsConfig.remoteUserPassword)
//		if err != nil {
//			return err
//		}
//	}
//	var bridgeNumber string
//	installedBridgesLen := len(installedBridges)
//	if installedBridgesLen == 0 {
//		return errors.New("no bridge found")
//	}
//	for _, network := range NETWORKS {
//		if installedBridgesLen == 1 {
//			vac.vaBridgesMap[network][0] = installedBridges[0]
//		} else {
//			for {
//				vac.ui.SetSize(10+installedBridgesLen, 55)
//				vac.ui.SetLabel(fmt.Sprintf("Choose appropriate bridge for \"%s\" network:", network))
//				bridgeNumber = vac.ui.MenuSlice(installedBridgesLen, installedBridges)
//				if bridgeNumber != "" {
//					break
//				}
//			}
//			bridgeNumberInt, err := strconv.Atoi(bridgeNumber)
//			if err != nil {
//				return err
//			}
//			vac.vaBridgesMap[network][0] = installedBridges[bridgeNumberInt-1]
//		}
//	}
//	return nil
//}

//// Confirmation
//func (vac *vaConfig) confirmation() {
//	var buf bytes.Buffer
//	buf.WriteString("Confirmation\n")
//	buf.WriteString("---------------")
//	buf.WriteString("\nEnvironment => " + ENVIRONMENT)
//	height := 19
//	if vac.vaInstallType == LOCAL {
//		height = 17
//		buf.WriteString("\nDeployment type => local")
//	} else {
//		buf.WriteString("\nDeployment type => remote")
//	}
//	buf.WriteString("\nPath to QEMU/KVM emulator => " + vac.vaKvmEmulatorPath[0])
//	buf.WriteString("\nDefault appliance name => " + vac.vaName[0])
//	buf.WriteString("\nNew appliance name => " + vac.vaName[1])
//	buf.WriteString("\nTarget image name and location => " + vac.vaExportDir[0] + "/" + vac.vaName[1] + ".img")
//	for _, network := range NETWORKS {
//		buf.WriteString(fmt.Sprintf("\n\"%s\" network  bridge => %s ", network, vac.vaBridgesMap[network][0]))
//	}
//	if vac.vaInstallType == REMOTE {
//		buf.WriteString("\nRemote host IP => " + vac.rsConfig.remoteIpAddress)
//		buf.WriteString("\nRemote user =>  " + vac.rsConfig.remoteUser)
//	}
//	buf.WriteString("\n\nPress <OK> to proceed or <CTRL+C> to exit")
//	vac.ui.SetSize(height, 100)
//	vac.ui.Msgbox(buf.String())
//}
