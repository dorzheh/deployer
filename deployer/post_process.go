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
	duration, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}
	progressBarTitle := c.VaName + " post-processing progress"
	progressBarMsg := "\n\nPlease wait..."
	return c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
}
