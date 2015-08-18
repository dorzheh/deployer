package deployer

import (
	"errors"
	"github.com/dorzheh/deployer/ui/dialog_ui"
	"os"
)

var SkipStep = errors.New("skip step")

type uiCtrl struct {
	funcs []func() error
}

func NewUiStepsController() *uiCtrl {
	return &uiCtrl{make([]func() error, 0)}
}

func (c *uiCtrl) RegisterSteps(fs ...func() error) {
	for _, f := range fs {
		c.funcs = append(c.funcs, f)
	}
}

func (c *uiCtrl) RunSteps() error {
	stepMoveBack := false
	for i := 0; i < len(c.funcs); {
		err := c.funcs[i]()
		if err != nil {
			if err == SkipStep {
				if stepMoveBack {
					if i > 0 {
						i--
					}
					stepMoveBack = false
				} else {
					i++
				}
				continue
			}
			switch err.Error() {
			case dialog_ui.DialogExit:
				os.Exit(1)
			case dialog_ui.DialogMoveBack:
				if i > 0 {
					i--
				}
				stepMoveBack = true
				continue
			default:
				return err
			}
		}
		stepMoveBack = false
		i++
	}
	return nil
}
