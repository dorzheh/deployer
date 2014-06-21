package config

import (
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/comm/ssh"
	"github.com/dorzheh/infra/utils/lshw"
	"github.com/dorzheh/mxj"
)

func GetHwInfo(lshwpath string, lshwconf *lshw.Config, sshconf *ssh.Config) (mxj.Map, error) {
	l, err := lshw.New(lshwpath, lshwconf)
	if err != nil {
		return nil, err
	}
	run := deployer.RunFunc(sshconf)
	out, err := run(l.Cmd())
	if err != nil {
		return nil, err
	}
	m, err := mxj.NewMapXml([]byte(out))
	if err != nil {
		return nil, err
	}
	return m, nil
}
