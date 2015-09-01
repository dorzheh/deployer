package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/deployer"
	libvirt_kvm "github.com/dorzheh/deployer/example/myproduct/env/libvirt/kvm"
	"github.com/dorzheh/deployer/example/myproduct/env/openxen"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/infra/utils/archutils"
)

const arch = "x86_64"

var (
	defaultProductName = ""
	rootDir            = ""
)

// Initialize the stuff before the main() is executed.
// Setting default product name , path to the installer ,
// adding internal path to bin
func init() {
	var absPath string
	absPath, _ = filepath.Abs(os.Args[0])
	var pathArr []string = strings.Split(absPath, "/")
	rootDir = strings.Join(pathArr[0:len(pathArr)-4], "/")
	newPath := "$PATH:" + rootDir + "/install/" + arch + "/bin"
	os.Setenv("PATH", (os.ExpandEnv(newPath)))
	buf, err := ioutil.ReadFile(filepath.Join(rootDir, ".product"))
	if err != nil {
		panic(err)
	}

	defaultProductName = strings.TrimSpace(strings.Split(string(buf), "=")[1])
}

func main() {
	ui := dialog_ui.NewDialogUi()
	ui.Shadow(false)
	ui.SetCancelLabel("Exit")
	gui.UiValidateUser(ui, 0)
	gui.UiWelcomeMsg(ui, defaultProductName)
	gui.UiEulaMsg(ui, filepath.Join(rootDir, ".EULA"))
	data := &deployer.CommonData{
		RootDir:          rootDir,
		RootfsMp:         filepath.Join(rootDir, "rootfs_mnt"),
		DefaultExportDir: rootDir,
		VaName:           defaultProductName,
		Arch:             arch,
		Ui:               ui,
	}

	if err := archutils.Extract(filepath.Join(rootDir, "comp/env.tgz"), filepath.Join(rootDir, "comp")); err != nil {
		ui.Output(dialog_ui.Error, err.Error())
	}

	gui.UiDeploymentResult(ui, data.VaName+" installation completed successfully",
		gui.UiSelectEnv(data, []string{"Libvirt(KVM)", "OpenXen"},
			[]deployer.FlowCreator{new(libvirt_kvm.FlowCreator), new(openxen.FlowCreator)}))
}
