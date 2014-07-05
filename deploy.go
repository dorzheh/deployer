package deployer

import (
	"github.com/dorzheh/deployer/deployer"
)

func Deploy(c *deployer.CommonData, f deployer.FlowCreator) error {
	if err := f.CreateConfig(c); err != nil {
		return err
	}

	builders, err := f.CreateBuilders()
	if err != nil {
		return err
	}

	artifacts, err := deployer.BuildProgress(c, builders)
	if err != nil {
		return err
	}

	prov, err := f.CreateProvisioner()
	if err != nil {
		return err
	}

	if prov != nil {
		if artifacts, err = deployer.ProvisionProgress(c, prov, artifacts); err != nil {
			return err
		}
	}

	post, err := f.CreatePostProcessor()
	if err != nil {
		return err
	}
	if post != nil {
		return deployer.PostProcessProgress(c, post, artifacts)
	}
	return nil
}
