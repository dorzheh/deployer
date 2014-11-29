package deployer

import (
	"runtime"
)

// BuildProgress is responsible for running appropriate builders
// and representing a progress bar providing information about the build progress.
func BuildProgress(c *CommonData, builders []Builder) (artifacts []Artifact, err error) {
	if c.Ui == nil {
		return Build(builders)
	}
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		artifacts, err = Build(builders)
		errChan <- err
	}()

	progressBarTitle := c.VaName + " installation in progress (artifacts building stage)"
	progressBarMsg := "\n\nPlease wait..."
	err = c.Ui.Progress(progressBarTitle, progressBarMsg, c.Ui.Pb.Sleep(), c.Ui.Pb.Step(), errChan)
	return
}

// buildResult contains result of a build.
type buildResult struct {
	artifact Artifact
	err      error
}

// Build iterates over a slice of builders and runs
// each builder in a separated goroutine.
// Returns a slice of artifacts.
func Build(builders []Builder) ([]Artifact, error) {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	var artifacts []Artifact
	ch := make(chan *buildResult, len(builders))
	defer close(ch)

	for _, b := range builders {
		go func(b Builder) {
			artifact, err := b.Run()
			// Forwards created artifact to the channel.
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
