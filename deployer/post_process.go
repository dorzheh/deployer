package deployer

import (
	"github.com/dorzheh/deployer/utils"
)

// PostProcessProgress is responsible for representing a progress
// during post-processing of appropriate artifact.
func PostProcessProgress(c *CommonData, p PostProcessor, artifacts []Artifact) error {
	if c.Ui == nil {
		return p.PostProcess(artifacts)
	}

	errChan := make(chan error)
	defer close(errChan)
	go func() {
		errChan <- p.PostProcess(artifacts)
	}()

	progressBarTitle := c.VaName + " installation in progress (post-processing stage)"
	progressBarMsg := "\n\nPlease wait..."
	if err := c.Ui.Progress(progressBarTitle, progressBarMsg, c.Ui.Pb.Sleep(), c.Ui.Pb.Step(), errChan); err != nil {
		return utils.FormatError(err)
	}
	return nil
}
