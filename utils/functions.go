package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	sshconf "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/ssh"
)

const (
	PRE_SCRIPTS = iota
	POST_SCRIPTS
)

type ConnFuncAlias func(*sshconf.Config) (*ssh.SshConn, error)

func ConnFunc(config *sshconf.Config) func() (*ssh.SshConn, error) {
	return func() (*ssh.SshConn, error) {
		c, err := ssh.NewSshConn(config)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

// RunFunc is a generic solution for running appropriate commands
// on local or remote host
func RunFunc(config *sshconf.Config) func(string) (string, error) {
	if config == nil {
		return func(command string) (string, error) {
			var stderr bytes.Buffer
			var stdout bytes.Buffer
			c := exec.Command("/bin/bash", "-c", command)
			c.Stderr = &stderr
			c.Stdout = &stdout
			if err := c.Start(); err != nil {
				return "", err
			}
			if err := c.Wait(); err != nil {
				return "", fmt.Errorf("executing %s  : %s [%s]", command, stderr.String(), err)
			}
			return strings.TrimSpace(stdout.String()), nil
		}
	}
	return func(command string) (string, error) {
		c, err := ssh.NewSshConn(config)
		if err != nil {
			return "", err
		}
		defer c.ConnClose()
		outstr, errstr, err := c.Run(command)
		if err != nil {
			return "", fmt.Errorf("executing %s : %s [%s]", command, errstr, err)
		}
		return strings.TrimSpace(outstr), nil
	}
}

// ProcessTemplate is responsible for writing appapropriate user data to
// any metadata needed for deploying appliance at a given environment
func ProcessTemplate(str string, userData interface{}) ([]byte, error) {
	t, err := template.New(time.Now().String()).Parse(str)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, userData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WaitForResult waits for multyple goroutines to finish
// and returns result
func WaitForResult(ch <-chan error, num int) error {
	for i := 0; i < num; i++ {
		select {
		case result := <-ch:
			if result != nil {
				return result
			}
		}
	}
	return nil
}

// RunPrePostScripts gets a value representing the pre or post action
// and executes the scripts in a numeric order.Current valuable usage -
// executing deployer locally on a cloud instance
// The scripts must contain the following prefix : [0-9]+_
// Example: 02_clean
//returns error or nil
func RunPrePostScripts(pathToPlatformDir string, preOrPost uint8) error {
	d, err := os.Stat(pathToPlatformDir)
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return fmt.Errorf("%s is not directory", d.Name())
	}
	var pathToScriptsDir string
	switch preOrPost {
	case PRE_SCRIPTS:
		pathToScriptsDir = pathToPlatformDir + "/pre-deploy-scripts"
	case POST_SCRIPTS:
		pathToScriptsDir = pathToPlatformDir + "/post-deploy-scripts"
	default:
		return errors.New("unknown stage")
	}
	fd, err := os.Stat(pathToScriptsDir)
	if err != nil {
		return err
	}

	if !fd.IsDir() {
		return fmt.Errorf("%s is not directory", fd.Name())
	}
	var scriptsSlice []string
	//find mapped loop device partition , create appropriate mount point for each partition
	err = filepath.Walk(pathToScriptsDir, func(scriptName string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if found, _ := regexp.MatchString("[0-9]+_", scriptName); found {
				scriptsSlice = append(scriptsSlice, scriptName)
			}
		}
		return nil
	})
	sort.Strings(scriptsSlice)
	for _, file := range scriptsSlice {
		if err := exec.Command(file).Run(); err != nil {
			return err
		}
	}
	return err
}
