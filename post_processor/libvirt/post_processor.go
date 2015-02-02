package libvirt

import (
	"strings"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
)

type PostProcessor struct {
	Driver      *Driver
	StartDomain bool
}

func (p *PostProcessor) PostProcess(artifacts []deployer.Artifact) error {
	for _, a := range artifacts {
		switch a.(type) {
		case *deployer.CommonArtifact:
			if a.GetType() == deployer.MetadataArtifact {
				defer a.Destroy()
				domain := strings.Split(a.GetName(), "-metadata")[0]
				if err := p.Driver.DefineDomain(a.GetPath()); err != nil {
					return utils.FormatError(err)
				}
				if err := p.Driver.SetAutostart(domain); err != nil {
					return utils.FormatError(err)
				}
				if p.StartDomain {
					if err := p.Driver.StartDomain(domain); err != nil {
						return nil
					}
				}
			}
		}
	}
	return nil
}
