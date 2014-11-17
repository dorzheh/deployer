package utils

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
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

// UploadBinaries is intended to create a temporary directory on a remote server,
// upload binaries to the temporary location and return path to the directory.
// The tempoary directory will be removed as soon as the function exits
func UploadBinaries(conf *sshconf.Config, pathbins ...string) (string, error) {
	c, err := ssh.NewSshConn(conf)
	if err != nil {
		return "", err
	}
	defer c.ConnClose()

	dir, errout, err := c.Run("mktemp -d --suffix _deployer_bin")
	if err != nil {
		return "", fmt.Errorf("%s [%v]", errout, err)
	}

	dir = strings.TrimSpace(dir)

	for _, src := range pathbins {
		dst := filepath.Join(dir, filepath.Base(src))
		if err := c.Upload(src, dst); err != nil {
			return "", err
		}
		if _, errout, err := c.Run("chmod 755 " + dst); err != nil {
			return "", fmt.Errorf("%s [%v]", errout, err)
		}
	}
	return dir, nil
}

func ParseXMLFile(xmlpath string, data interface{}) (interface{}, error) {
	fb, err := ioutil.ReadFile(xmlpath)
	if err != nil {
		return nil, err
	}
	return ParseXMLBuff(fb, data)
}

func ParseXMLBuff(fb []byte, data interface{}) (interface{}, error) {
	buf := bytes.NewBuffer(fb)
	decoded := xml.NewDecoder(buf)
	if err := decoded.Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}
