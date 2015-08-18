package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/example/myproduct"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/ui/dialog_ui"
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
	gui.UiDeploymentResult(ui, data.VaName+" installation completed successfully", myproduct.Deploy(data))
}
