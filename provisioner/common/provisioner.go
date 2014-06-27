package deployer

import (
	"fmt"
	"path/filepath"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/comm/ssh"
)

type Provisioner struct {
	ConnFunc func() (*ssh.SshConn, error)
	DstDir   string
}

func (p *Provisioner) Provision(artifacts []deployer.Artifact) (artfcts []deployer.Artifact, err error) {
	conn, err := p.ConnFunc()
	if err != nil {
		return
	}
	defer conn.ConnClose()

	for _, srcArtifact := range artifacts {
		switch srcArtifact.(type) {
		case *deployer.LocalArtifact:
			if err = conn.Upload(srcArtifact.GetPath(), p.DstDir); err != nil {
				return
			}
			cmd := fmt.Sprintf("tar xfvzp %s -C %s", filepath.Join(p.DstDir,
				filepath.Base(srcArtifact.GetPath()), p.DstDir))
			if _, errout, e := conn.Run(cmd); e != nil {
				err = fmt.Errorf("%s [%s]", errout, e)
				return
			}
			newArtifact := &deployer.RemoteArtifact{
				Name: srcArtifact.GetName(),
				Path: filepath.Join(p.DstDir, srcArtifact.GetName()),
				Type: srcArtifact.GetType(),
			}
			if err = srcArtifact.Destroy(); err != nil {
				return
			}
			artfcts = append(artfcts, newArtifact)
		}
	}
	return
}