package libvirt

import (
	"github.com/dorzheh/deployer/deployer"
)

type PostProcessor struct {
	Driver      Driver
	DomName     string
	StartDomain bool
}

func (p *PostProcessor) Process(artifacts []deployer.Artifact) error {
	for _, a := range artifacts {
		if a.GetType() == deployer.MetadataArtifact {
			if err := p.Driver.DefineDomain(a.GetPath()); err != nil {
				return err
			}
			if err := p.Driver.SetAutostart(p.DomName); err != nil {
				return err
			}
			if p.StartDomain {
				if err := p.Driver.StartDomain(p.DomName); err != nil {
					return nil
				}
			}
		}
	}
	return nil
}
