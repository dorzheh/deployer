package dialog_ui

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	. "github.com/dorzheh/go-dialog"
)

const (
	Success      = "SUCCESS"
	Error        = "ERROR"
	Warning      = "WARNING"
	Notification = "NOTIFICATION"
	None         = ""
)

type Pb struct {
	sleep time.Duration
	step  int
}

func (p *Pb) SetSleep(s string) (err error) {
	p.sleep, err = time.ParseDuration(s)
	return
}

func (p *Pb) SetStep(s int) {
	p.step = s
}

func (p *Pb) Sleep() time.Duration {
	return p.sleep
}

func (p *Pb) Step() int {
	return p.step
}

func (p *Pb) IncreaseSleep(s string) error {
	sleep, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	p.sleep += sleep
	return nil
}

func (p *Pb) IncreaseStep(s int) {
	p.step += s
}

func (p *Pb) DecreaseStep(s int) {
	p.step -= s
}

type DialogUi struct {
	*Dialog
	Pb *Pb
}

func NewDialogUi() *DialogUi {
	return &DialogUi{New(CONSOLE, 0), &Pb{0, 0}}
}

///// Functions providing verification services /////

// ErrorOutput gets dialog session , error string and height/width
// It prints out the error output inside dialog inforbox.
// The session is terminated with exit 1
func (ui *DialogUi) ErrorOutput(err string, height, widthOffset int) {
	ui.SetSize(height, len(err)+widthOffset)
	ui.Infobox("\n" + Error + ": " + err)
	os.Exit(1)
}

// Output gets dialog session and a msg string and height/width
// It prints out appropriate output inside dialog inforbox.
func (ui *DialogUi) Output(ntype string, msg string, height, widthOffset int) {
	if ntype == Notification || ntype == "" {
		ui.SetSize(height, widthOffset)
		ui.Msgbox(msg)
	} else {
		ui.SetSize(height, len(msg)+widthOffset)
		ui.Msgbox("\n" + ntype + ": " + msg)
	}
}

///// Functions for the progress bar implementation /////

// WaitForCmdToFinish prints a progress bar upon a command execution
// It gets a dialog session, command to execute,
// title for progress bar and the time duration
// Returns error
func (ui *DialogUi) WaitForCmdToFinish(cmd *exec.Cmd, title, msg string, step int, duration time.Duration) error {
	// execute the command in a background
	err := cmd.Start()
	if err != nil {
		return err
	}
	// allocate a channel
	done := make(chan error)
	go func() {
		// wait in background until the command has make it's job
		done <- cmd.Wait()
	}()
	// show progress bar for a while
	return ui.Progress(title, msg, duration, step, done)
}

// Progress implements a progress bar
// Returns error or nil
func (ui *DialogUi) Progress(title, pbMsg string, duration time.Duration, step int, done chan error) error {
	defaultWidth := 50
	titleWidth := len(title) + 4
	msgWidth := len(pbMsg) + 4
	var newWidth int
	if titleWidth > msgWidth {
		newWidth = titleWidth
	} else {
		newWidth = msgWidth
	}
	if defaultWidth > newWidth {
		newWidth = defaultWidth
	}
	ui.SetTitle(title)
	ui.SetSize(8, newWidth)
	pb := ui.Progressbar()
	var interval int = 0
	for {
		select {
		// wait for result
		case result := <-done:
			if result != nil {
				return result
			}
			// we are finished - 100% done
			pb.Step(100, "\n\nSUCCESS!")
			ui.SetSize(6, 15)
			finalSleep, err := time.ParseDuration("1s")
			if err != nil {
				return err
			}
			time.Sleep(finalSleep)
			return nil
		default:
			if interval < 100 {
				interval += step
			}
			if interval > 100 {
				interval = 100
			}
			pb.Step(interval, pbMsg)
			time.Sleep(duration)
		}
	}
	return nil
}

// Wait communicates with a progress bar while a given function is executed
// Returns error or nil
func (ui *DialogUi) Wait(msg string, pause time.Duration, done chan error) error {
	ui.SetSize(6, 55)
	ui.Infobox(msg)
	for {
		select {
		// wait for result
		case result := <-done:
			return result
		default:
			time.Sleep(pause)
		}
	}
	return nil
}

// GetPathToFileFromInput uses a dialog session for getting path to a file
func (ui *DialogUi) GetPathToFileFromInput(backtitle string) string {
	var result string
	for {
		ui.SetBackTitle(backtitle)
		ui.SetSize(10, 50)
		result = ui.Fselect("/")
		if result != "" {
			stat, err := os.Stat(result)
			if err == nil && !stat.IsDir() {
				break
			}
		}
	}
	return result
}

// GetPathToDirFromInput uses a dialog session for getting path to a directory to upload
func (ui *DialogUi) GetPathToDirFromInput(backtitle, defaultDir string) string {
	var result string
	for {
		ui.SetBackTitle(backtitle)
		ui.SetSize(10, 50)
		result = ui.Dselect(defaultDir)
		if result != "" {
			stat, err := os.Stat(result)
			if err == nil && stat.IsDir() {
				break
			}
		}
	}
	return result
}

// GetIpFromInput uses a dialog session for reading IP from user input
// Returns host IP (remote or local)
func (ui *DialogUi) GetIpFromInput(labelMsg string) string {
	var ipAddr string
	width := len(labelMsg) + 5
	for {
		ui.SetSize(8, width)
		ui.SetLabel(labelMsg)
		ipAddr = ui.Inputbox("")
		// validate the IP
		if net.ParseIP(ipAddr) == nil {
			ui.SetSize(5, 20)
			ui.Msgbox("Invalid IP!")
			continue
		}
		break
	}
	return ipAddr
}

// GetFromInput uses a dialog session for reading from stdin
// Returns user input
func (ui *DialogUi) GetFromInput(labelMsg string, defaultInput string) string {
	var input string
	for {
		ui.SetSize(8, len(labelMsg)+5)
		ui.SetLabel(labelMsg)
		input = ui.Inputbox(defaultInput)
		if input != "" {
			break
		}
	}
	return input
}

//GetPasswordFromInput uses a dialog session for reading user password from user input
//Returns password string
func (ui *DialogUi) GetPasswordFromInput(host, user string, confirm bool) (passwd1 string) {
	for {
		msg := fmt.Sprintf("\"%s\" password on the host %s", user, host)
		for {
			ui.SetSize(8, len(msg)+5)
			ui.SetLabel(msg)
			passwd1 = ui.Passwordbox(true)
			if passwd1 != "" {
				return
			}
		}
		if confirm {
			var passwd2 string
			msg = "Password confirmation for user \"" + user + "\""
			for {
				ui.SetSize(8, len(msg)+5)
				ui.SetLabel(msg)
				passwd2 = ui.Passwordbox(true)
				if passwd2 != "" {
					break
				}
			}
			if passwd1 == passwd2 {
				return
			}
		}
	}
	return
}
