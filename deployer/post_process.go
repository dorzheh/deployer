package deployer

import "time"

func PostProcessProgress(c *CommonData, p PostProcessor, artifacts []Artifact) error {
	if c.Ui == nil {
		return p.PostProcess(artifacts)
	}

	errChan := make(chan error)
	defer close(errChan)
	go func() {
		errChan <- p.PostProcess(artifacts)
	}()
	duration, err := time.ParseDuration("3s")
	if err != nil {
		return err
	}
	progressBarTitle := c.VaName + " installation in progress (post-processing stage)"
	progressBarMsg := "\n\nPlease wait..."
	return c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
}
