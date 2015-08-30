package xen_xl

import (
	"regexp"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/drivers/env_driver/openxen/xen_xl"
	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

type PostProcessor struct {
	driver      *xen_xl.Driver
	startDomain bool
}

func NewPostProcessor(sshconf *ssh.Config, startDomain bool) *PostProcessor {
	p := new(PostProcessor)
	p.driver = xen_xl.NewDriver(sshconf)
	p.startDomain = startDomain
	return p
}

func (p *PostProcessor) PostProcess(artifacts []deployer.Artifact) error {
	for _, a := range artifacts {
		switch a.(type) {
		case *deployer.CommonArtifact:
			if a.GetType() == deployer.MetadataArtifact {
				out, err := p.driver.Run("cat " + a.GetPath())
				if err != nil {
					return utils.FormatError(err)
				}

				r, err := regexp.Compile(`\s*name\s*=\s*(\S+)`)
				if err != nil {
					return utils.FormatError(err)
				}

				domain := r.FindStringSubmatch(out)[1]
				configFile := "/etc/xen/" + domain + ".cfg"
				if _, err := p.driver.Run("mkdir -p /etc/xen/auto;cp " + a.GetPath() + " " + configFile); err != nil {
					return utils.FormatError(err)
				}
				if err := p.driver.SetAutostart(domain); err != nil {
					return utils.FormatError(err)
				}
				if p.startDomain {
					if err := p.driver.StartDomain(configFile); err != nil {
						return utils.FormatError(err)
					}
				}
				if err := a.Destroy(); err != nil {
					return utils.FormatError(err)
				}
			}
		}
	}
	return nil
}
