// Copyright 2012 Pavel Vershinin. All rights reserved.
// master-dev@inbox.ru
// Use of this source code is governed by a BSD-style

// Dmitry Orzhehovsky <dorzheh@gmail.com>
// Adding new functionality

package dialog

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	CONSOLE = "dialog"
	KDE     = "kdialog"
	GTK     = "gtkdialog"
	X       = "Xdialog"
	AUTO    = "auto"
)

const (
	DIALOG_ERR_CANCEL = "exit status 1"
	DIALOG_ERR_HELP   = "exit status 2"
	DIALOG_ERR_EXTRA  = "exit status 3"
)

type Dialog struct {
	environment string
	parentId    int
	title       string
	backtitle   string
	label       string
	height      int
	width       int
	left        int
	top         int
	shadow      bool
	helpButton  bool
	helpLabel   string
	extraLabel  string
	cancelLabel string
	okLabel     string
	beforeDtype []string
	beforeSize  []string
	afterSize   []string
}

func New(environment string, parentId int) *Dialog {
	var err error
	var res = new(Dialog)
	if environment == AUTO || environment == "" {
		for _, pkg := range []string{KDE, GTK, X, CONSOLE} {
			_, err = exec.LookPath(pkg)
			if err == nil {
				res.environment = pkg
				break
			}
		}
		if res.environment == "" {
			fmt.Println("Package not found!\nPlease install " + KDE + " or " + GTK + " or " + X + " or " + CONSOLE)
		}
	} else {
		_, err = exec.LookPath(environment)
		if err == nil {
			res.environment = environment
		} else {
			fmt.Println("Package not found!\nPlease install " + environment)
		}
	}

	if res.environment == "" {
		os.Exit(1)
	}

	res.parentId = parentId
	res.reset()
	return res
}

func (d *Dialog) Shadow(truefalse bool) {
	d.shadow = truefalse
}

func (d *Dialog) SetCancelLabel(label string) {
	d.cancelLabel = label
}

func (d *Dialog) SetSize(height int, width int) {
	d.height = height
	d.width = width
}

func (d *Dialog) SetTitle(title string) {
	d.title = title
}

func (d *Dialog) SetBackTitle(backtitle string) {
	d.backtitle = backtitle
}

func (d *Dialog) SetLabel(label string) {
	d.label = label
}

func (d *Dialog) SetOkLabel(label string) {
	d.okLabel = label
}

func (d *Dialog) HelpButton(truefalse bool) {
	d.helpButton = truefalse
}

func (d *Dialog) SetHelpLabel(label string) {
	d.helpLabel = label
}

func (d *Dialog) SetExtraLabel(label string) {
	d.extraLabel = label
}

func (d *Dialog) reset() {
	d.title = ""
	d.backtitle = ""
	d.label = ""
	d.okLabel = "OK"
	d.extraLabel = ""
	d.helpButton = false
	d.helpLabel = ""
	d.SetSize(0, 0)
	d.beforeDtype = []string{}
	d.afterSize = []string{}
	d.beforeSize = []string{}
}

func (d *Dialog) exec(dType string, allowLabel bool) (string, error) {
	var arg string
	cmd := exec.Command(d.environment)

	if d.okLabel != "" {
		cmd.Args = append(cmd.Args, "--ok-label", d.okLabel)
	}
	if d.extraLabel != "" {
		cmd.Args = append(cmd.Args, "--extra-button")
		cmd.Args = append(cmd.Args, "--extra-label", d.extraLabel)
	}
	if d.helpButton {
		cmd.Args = append(cmd.Args, "--help-button")
		if d.helpLabel != "" {
			cmd.Args = append(cmd.Args, "--help-label", d.helpLabel)
		}
	}
	if d.shadow == false {
		cmd.Args = append(cmd.Args, "--no-shadow")
	}
	if d.backtitle != "" {
		cmd.Args = append(cmd.Args, "--backtitle", d.backtitle)
	}
	if d.cancelLabel != "" {
		cmd.Args = append(cmd.Args, "--cancel-label", d.cancelLabel)
	}
	if d.title != "" {
		cmd.Args = append(cmd.Args, "--title", d.title)
	}
	for _, arg := range d.beforeDtype {
		cmd.Args = append(cmd.Args, arg)
	}
	cmd.Args = append(cmd.Args, "--"+dType)

	if allowLabel == true {
		cmd.Args = append(cmd.Args, d.label)
	}
	for _, arg = range d.beforeSize {
		cmd.Args = append(cmd.Args, arg)
	}
	if d.environment != KDE {
		cmd.Args = append(cmd.Args, strconv.Itoa(d.height))
		cmd.Args = append(cmd.Args, strconv.Itoa(d.width))
	}
	for _, arg = range d.afterSize {
		cmd.Args = append(cmd.Args, arg)
	}
	if d.environment == CONSOLE {
		cmd.Args = append(cmd.Args, "--stdout")
	} else {
		cmd.Args = append(cmd.Args, "--attach")
		cmd.Args = append(cmd.Args, strconv.Itoa(d.parentId))
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	d.reset()
	return strings.Trim(out.String(), "\r\n "), err
}

func (d *Dialog) Slider(min int, max int, step int) (int, error) {
	d.afterSize = append(d.afterSize, strconv.Itoa(min))
	d.afterSize = append(d.afterSize, strconv.Itoa(max))
	d.afterSize = append(d.afterSize, strconv.Itoa(step))
	val, err := d.exec("slider", true)
	res, err1 := strconv.Atoi(val)
	if err1 != nil {
		return res, err1
	}
	return res, err
}

func (d *Dialog) Passivepopup(text string, timeout int) {
	d.afterSize = append(d.afterSize, text)
	d.afterSize = append(d.afterSize, strconv.Itoa(timeout))
	d.exec("passivepopup", false)
}

func (d *Dialog) Geticon() string {
	val, _ := d.exec("geticon", false)
	return val
}

func (d *Dialog) Getcolor() string {
	val, _ := d.exec("getcolor", false)
	return val
}

func (d *Dialog) Combobox(item ...string) (string, error) {
	var command string
	if d.environment == CONSOLE {
		d.afterSize = append(d.afterSize, "0")
		for _, param := range item {
			d.afterSize = append(d.afterSize, param)
			d.afterSize = append(d.afterSize, param)
		}
		command = "menu"
	} else {
		for _, param := range item {
			d.afterSize = append(d.afterSize, param)
		}
		command = "combobox"
	}
	return d.exec(command, true)
}

func (d *Dialog) Calendar(date time.Time) (string, error) {
	d.afterSize = append(d.afterSize, date.Format("2006"))
	d.afterSize = append(d.afterSize, date.Format("01"))
	d.afterSize = append(d.afterSize, date.Format("02"))
	return d.exec("calendar", true)
}

func (d *Dialog) Checklist(listHeight int, tagItemStatus ...string) ([]string, error) {
	var list []string
	d.afterSize = append(d.afterSize, strconv.Itoa(listHeight))
	for _, param := range tagItemStatus {
		d.afterSize = append(d.afterSize, param)
	}
	str, err := d.exec("checklist", true)
	for _, item := range strings.Split(str, " ") {
		list = append(list, strings.Replace(item, "\"", "", -1))
	}
	return list, err
}

func (d *Dialog) Mixedform(title string, tagItemStatus ...string) ([]string, error) {
	var list []string
	d.afterSize = append(d.afterSize, "0")
	for _, param := range tagItemStatus {
		d.afterSize = append(d.afterSize, param)
	}

	d.beforeSize = append(d.beforeSize, title)
	str, err := d.exec("mixedform", false)
	for _, item := range strings.SplitAfter(str, "\n") {
		list = append(list, strings.TrimSpace(item))
	}
	return list, err
}

func (d *Dialog) Fselect(filepath string) (string, error) {
	d.beforeSize = append(d.beforeSize, filepath)
	var command string
	if d.environment == KDE {
		command = "getopenfilename"
	} else {
		command = "fselect"
	}
	return d.exec(command, false)
}

func (d *Dialog) Infobox(text string) {
	d.beforeSize = append(d.beforeSize, text)
	var command string
	if d.environment == KDE {
		command = "msgbox"
	} else {
		command = "infobox"
	}
	d.exec(command, false)
}

func (d *Dialog) Inputbox(value string) (string, error) {
	d.afterSize = append(d.afterSize, value)
	return d.exec("inputbox", true)
}

func (d *Dialog) Inputmenu(menuHeight int, tagItem ...string) ([]string, error) {
	d.afterSize = append(d.afterSize, strconv.Itoa(menuHeight))
	for _, param := range tagItem {
		d.afterSize = append(d.afterSize, param)
	}

	var command string
	if d.environment == KDE {
		command = "menu"
	} else {
		command = "inputmenu"
	}

	val, err := d.exec(command, true)
	res := strings.Split(val, "\n")
	return res, err
}

func (d *Dialog) Menu(menuHeight int, tagItem ...string) (string, error) {
	d.afterSize = append(d.afterSize, strconv.Itoa(menuHeight))
	for _, param := range tagItem {
		d.afterSize = append(d.afterSize, param)
	}
	return d.exec("menu", true)
}

func (d *Dialog) Msgbox(text string) {
	d.beforeSize = append(d.beforeSize, text)
	d.exec("msgbox", false)
}

func (d *Dialog) Passwordbox(insecure bool) (string, error) {
	var command string
	if d.environment == KDE {
		command = "password"
	} else {
		if insecure {
			d.beforeDtype = append(d.beforeDtype, "--insecure")
		}
		d.afterSize = append(d.afterSize, "")
		command = "passwordbox"
	}
	return d.exec(command, true)
}

func (d *Dialog) Pause(seconds int) {
	if d.environment == KDE {
		var percent = int(100 / seconds)
		var p = d.Progressbar()
		p.Step(100, "Pause "+strconv.Itoa(seconds)+" seconds")
		for i := seconds; i > 0; i-- {
			p.Step(int(percent*i), "Pause "+strconv.Itoa(i)+" seconds")
			time.Sleep(1 * time.Second)
		}
		p.Close()
	} else {
		d.afterSize = append(d.afterSize, strconv.Itoa(seconds))
		d.exec("pause", true)
	}
}

func (d *Dialog) Textbox(filepath string) error {
	d.beforeSize = append(d.beforeSize, filepath)
	_, err := d.exec("textbox", false)
	return err
}

func (d *Dialog) Timebox(date time.Time) (string, error) {
	d.afterSize = append(d.afterSize, date.Format("15"))
	d.afterSize = append(d.afterSize, date.Format("04"))
	d.afterSize = append(d.afterSize, date.Format("05"))
	return d.exec("timebox", true)
}

func (d *Dialog) Yesno() bool {
	if _, err := d.exec("yesno", true); err != nil {
		if err.Error() == DIALOG_ERR_CANCEL {
			return false
		}
	}
	return true
}

func (d *Dialog) Radiolist(listHeight int, tagItemStatus ...string) (string, error) {
	d.afterSize = append(d.afterSize, strconv.Itoa(listHeight))
	for _, param := range tagItemStatus {
		d.afterSize = append(d.afterSize, param)
	}
	val, err := d.exec("radiolist", true)
	str := strings.Replace(val, "\"", "", -1)
	return str, err
}

func (d *Dialog) Dselect(dirpath string) (string, error) {
	d.beforeSize = append(d.beforeSize, dirpath)
	var command string
	if d.environment == KDE {
		return "", nil
	}
	command = "dselect"
	return d.exec(command, false)
}

type progress struct {
	id          []string
	environment string
	label       string
	title       string
	shadow      bool
	height      int
	width       int
}

func (d *Dialog) Progressbar() *progress {
	var out []byte
	var id []string
	if d.environment == KDE {
		out, _ = exec.Command("kdialog", "--progressbar", "Initializing", "100", "--title", d.title).Output()
		id = strings.Split(strings.Trim(string(out), " \n\r"), " ")
	} else {
		cmd := exec.Command(d.environment)
		if d.shadow == false {
			cmd.Args = append(cmd.Args, "--no-shadow")
		}

		cmd.Args = append(cmd.Args, "--title", d.title, "--gauge", d.label, strconv.Itoa(d.height), strconv.Itoa(d.width), "0", "--stdout")
		cmd.Run()
	}
	var res = new(progress)
	res.id = id
	res.label = d.label
	res.environment = d.environment
	res.shadow = d.shadow
	res.height = d.height
	res.width = d.width
	res.title = d.title
	return res
}

func (p *progress) Step(percent int, newLabel string) {
	if newLabel == "" {
		newLabel = p.label
	}
	if p.environment == KDE {
		exec.Command("qdbus", p.id[0], p.id[1], "setLabelText", newLabel).Run()
		exec.Command("qdbus", p.id[0], p.id[1], "Set", "", "value", strconv.Itoa(percent)).Run()
	} else {
		cmd := exec.Command(p.environment)
		if p.shadow == false {
			cmd.Args = append(cmd.Args, "--no-shadow")
		}

		cmd.Args = append(cmd.Args, "--title", p.title, "--gauge", newLabel, strconv.Itoa(p.height), strconv.Itoa(p.width), strconv.Itoa(percent), "--stdout")
		cmd.Run()
	}
}

func (p *progress) Close() {
	if p.environment == KDE {
		exec.Command("qdbus", p.id[0], p.id[1], "close").Run()
	}
	p = nil
}
