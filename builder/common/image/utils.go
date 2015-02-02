package image

import (
	"bufio"
	"os"
	"strings"

	"github.com/dorzheh/deployer/utils"
)

// isMounted parses /proc/mounts and looking for
// appropriate entry representing device mapper.
// If the mapper is found - return true,
// otherwise return false
func isMounted(device string) (bool, error) {
	fh, err := os.Open("/proc/mounts")
	if err != nil {
		return false, utils.FormatError(err)
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		slice := strings.Split(line, " ")
		if slice[0] == device {
			return true, nil
		}
	}
	return false, nil
}
