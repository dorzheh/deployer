package deployer

import (
	"runtime"
	"time"
)

func BuildProgress(c *CommonData, builders []Builder) (artifacts []Artifact, err error) {
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		artifacts, err = Build(builders)
		if err != nil {
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
	progressBarTitle := c.VaName + " deployment progress"
	progressBarMsg := "\n\nPlease wait..."
	err = c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
	return
}

func Build(builders []Builder) (artifacts []Artifact, err error) {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	errCh := make(chan error, runtime.NumCPU())
	defer close(errCh)

	var artifact Artifact
	for _, b := range builders {
		go func() {
			if artifact, err = b.Run(); err != nil {
				errCh <- err
			}
			artifacts = append(artifacts, artifact)
		}()
	}
	if err = WaitForResult(errCh, len(builders)); err != nil {
		return
	}
	return
}

func WaitForResult(ch <-chan error, num int) error {
	for i := 0; i < num; i++ {
		select {
		case result := <-ch:
			if result != nil {
				return result
			}
		}
	}
	return nil
}
