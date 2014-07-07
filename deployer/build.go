package deployer

import (
	"runtime"
	"time"
)

func BuildProgress(c *CommonData, builders []Builder) (artifacts []Artifact, err error) {
	if c.Ui == nil {
		return Build(builders)
	}
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
	progressBarTitle := c.VaName + " artifacts building progress"
	progressBarMsg := "\n\nPlease wait..."
	err = c.Ui.Progress(progressBarTitle, progressBarMsg, duration, 15, errChan)
	return
}

type buildResult struct {
	artifact Artifact
	err      error
}

func Build(builders []Builder) ([]Artifact, error) {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	var artifacts []Artifact
	ch := make(chan *buildResult, len(builders))
	defer close(ch)

	for _, b := range builders {
		go func(b Builder) {
			artifact, err := b.Run()
			ch <- &buildResult{artifact, err}
		}(b)
	}
	for i := 0; i < len(builders); i++ {
		select {
		case result := <-ch:
			if result.err != nil {
				return nil, result.err
			}
			artifacts = append(artifacts, result.artifact)
		}
	}
	return artifacts, nil
}
