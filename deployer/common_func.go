package deployer

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/dorzheh/infra/comm/ssh"
)

func ConnFunc(config *ssh.Config) func() (*ssh.SshConn, error) {
	return func() (*ssh.SshConn, error) {
		c, err := ssh.NewSshConn(config)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

func RunFunc(config *ssh.Config) func(string) (string, error) {
	if config == nil {
		return func(command string) (string, error) {
			var stderr bytes.Buffer
			var stdout bytes.Buffer
			cmd := strings.Fields(command)
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stderr = &stderr
			c.Stdout = &stdout
			if err := c.Start(); err != nil {
				return "", err
			}
			if err := c.Wait(); err != nil {
				return "", fmt.Errorf("executing %s  : %s [%s]", cmd, stderr.String(), err)
			}
			return stdout.String(), nil
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
		return outstr, nil
	}
}

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
