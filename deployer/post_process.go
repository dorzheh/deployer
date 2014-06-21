package deployer

import "time"

func PostProcessProgress(c *CommonData, p PostProcessor, a []Artifact) error {
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		if err := p.Process(a); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()
	duration, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}
	progressBarTitle := c.VaName + " post-processing progress"
	progressBarMsg := "\n\nPlease wait..."
	return c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
}
