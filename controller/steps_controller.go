package controller

import (
	"errors"
	"github.com/dorzheh/deployer/ui/dialog_ui"
	"os"
)

var SkipStep = errors.New("skip step")

var steps []func() error

func init() {
	steps = make([]func() error, 0)
}

func RegisterSteps(fs ...func() error) {
	for _, f := range fs {
		steps = append(steps, f)
	}
}

func RunSteps() error {
	stepMoveBack := false
	for i := 0; i < len(steps); {
		err := steps[i]()
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
