package deployer

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"
	"time"

	sshconf "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/ssh"
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
