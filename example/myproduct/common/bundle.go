package common

import (
	"errors"

	gui "github.com/dorzheh/deployer/ui/dialog_ui"
)

func NameToType(ui *gui.DialogUi, configname string) error {
	switch configname {
	case "Bundle1":
		ui.Pb.SetSleep("30s")
		ui.Pb.SetStep(7)
		return nil
	case "Bundle2":
		ui.Pb.SetSleep("15s")
		ui.Pb.SetStep(10)
		return nil
	case "Bundle3":
		ui.Pb.SetSleep("30s")
		ui.Pb.SetStep(7)
		return nil
	case "Bundle4":
		ui.Pb.SetSleep("10s")
		ui.Pb.SetStep(10)
		return nil
	}

	return errors.New("unsupported configuration")
}
