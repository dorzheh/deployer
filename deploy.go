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
	builderArtifacts, err := deployer.BuildProgress(c, builders)
	if err != nil {
		return err
	}
	prov, err := f.CreateProvisioner()
	if err != nil {
		return err
	}
	var provisionArtifacts []deployer.Artifact
	if prov != nil {
		if provisionArtifacts, err = deployer.ProvisionProgress(c, prov, builderArtifacts); err != nil {
			return err
		}
	}
	post, err := f.CreatePostProcessor()
	if err != nil {
		return err
	}
	if post != nil {
		return deployer.PostProcessProgress(c, post, provisionArtifacts)
	}
	return nil
}
