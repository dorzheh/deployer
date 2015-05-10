package deployer

import (
	"runtime"
	"time"

	"github.com/dorzheh/deployer/utils"
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

	progressBarTitle := "Building artifacts"
	progressBarMsg := "\n" + c.VaName + " installation in progress.Please wait..."
	if err = c.Ui.Progress(progressBarTitle, progressBarMsg, c.Ui.Pb.Sleep(), c.Ui.Pb.Step(), errChan); err != nil {
		err = utils.FormatError(err)
	}
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
	dur, err := time.ParseDuration("1s")
	if err != nil {
		return nil, utils.FormatError(err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	var artifacts []Artifact
	ch := make(chan *buildResult, len(builders))

	for _, b := range builders {
		time.Sleep(dur)
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
				//defer close(ch)
				return nil, utils.FormatError(result.err)
			}
			artifacts = append(artifacts, result.artifact)
		}
	}
	return artifacts, nil
}
