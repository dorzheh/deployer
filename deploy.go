package deployer

import (
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
)

func Deploy(c *deployer.CommonData, f deployer.FlowCreator) error {
	if err := f.CreateConfig(c); err != nil {
		return utils.FormatError(err)
	}

	builders, err := f.CreateBuilders(c)
	if err != nil {
		return utils.FormatError(err)
	}

	artifacts, err := deployer.BuildProgress(c, builders)
	if err != nil {
		return utils.FormatError(err)
	}

	post, err := f.CreatePostProcessor(c)
	if err != nil {
		return utils.FormatError(err)
	}
	if post != nil {
		err := deployer.PostProcessProgress(c, post, artifacts)
		if err != nil {
			return utils.FormatError(err)
		}
	}
	return nil
}
