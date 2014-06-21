package deployer

import "time"

func ProvisionProgress(c *CommonData, p Provisioner, artifacts []Artifact) (artfcts []Artifact, err error) {
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		if artfcts, err = p.Provision(artifacts); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()
	var duration time.Duration
	duration, err = time.ParseDuration("10s")
	if err != nil {
		return
	}
	progressBarTitle := c.VaName + " provisioning progress"
	progressBarMsg := "\n\nPlease wait..."
	err = c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
	return
}
