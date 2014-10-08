package deployer

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
	return c.Ui.Progress(progressBarTitle, progressBarMsg, c.Ui.Pb.Sleep(), c.Ui.Pb.Step(), errChan)
}
