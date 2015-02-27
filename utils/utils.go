package utils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	sshconf "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/ssh"
)

// ProcessTemplate is responsible for writing appapropriate user data to
// any metadata needed for deploying appliance at a given environment
func ProcessTemplate(str string, userData interface{}) ([]byte, error) {
	t, err := template.New(time.Now().String()).Parse(str)
	if err != nil {
		return nil, FormatError(err)
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, userData); err != nil {
		return nil, FormatError(err)
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

// UploadBinaries is intended to create a temporary directory on a remote server,
// upload binaries to the temporary location and return path to the directory.
// The tempoary directory will be removed as soon as the function exits
func UploadBinaries(conf *sshconf.Config, pathbins ...string) (string, error) {
	c, err := ssh.NewSshConn(conf)
	if err != nil {
		return "", FormatError(err)
	}
	defer c.ConnClose()

	dir, errout, err := c.Run("mktemp -d --suffix _deployer_bin")
	if err != nil {
		return "", FormatError(fmt.Errorf("%s [%v]", errout, err))
	}

	dir = strings.TrimSpace(dir)

	for _, src := range pathbins {
		dst := filepath.Join(dir, filepath.Base(src))
		if err := c.Upload(src, dst); err != nil {
			return "", FormatError(err)
		}
		if _, errout, err := c.Run("chmod 755 " + dst); err != nil {
			return "", FormatError(fmt.Errorf("%s [%v]", errout, err))
		}
	}
	return dir, nil
}

func ParseXMLFile(xmlpath string, data interface{}) (interface{}, error) {
	fb, err := ioutil.ReadFile(xmlpath)
	if err != nil {
		return nil, FormatError(err)
	}
	return ParseXMLBuff(fb, data)
}

func ParseXMLBuff(fb []byte, data interface{}) (interface{}, error) {
	buf := bytes.NewBuffer(fb)
	decoded := xml.NewDecoder(buf)
	if err := decoded.Decode(data); err != nil {
		return nil, FormatError(err)
	}
	return data, nil
}

func FormatError(err error) error {
	pc, fn, line, _ := runtime.Caller(1)
	return fmt.Errorf("%v\nTRACE: %s[%s:%d]", err, runtime.FuncForPC(pc).Name(), filepath.Base(fn), line)
}
