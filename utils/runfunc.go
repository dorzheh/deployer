package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	sshconf "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/ssh"
)

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
				return "", FormatError(err)
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
			return "", FormatError(err)
		}
		defer c.ConnClose()

		var cmd string
		if strings.TrimSpace(config.User) == "root" {
			cmd = command
		} else {
			//IMPORTANT! Make sure that "Defaults !requiretty" is set in sudoers on remote system
			cmd = "sudo " + command
		}

		outstr, errstr, err := c.Run(cmd)
		if err != nil {
			return "", fmt.Errorf("executing %s : %s [%s]", cmd, errstr, err)
		}
		return strings.TrimSpace(outstr), nil
	}
}
