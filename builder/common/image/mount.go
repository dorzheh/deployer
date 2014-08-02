package image

import (
	"fmt"
	"strings"
)

const (
	mountinfoFormat = "%d %d %d:%d %s %s %s"
)

type procEntry struct {
	id, parent, major, minor int
	source, mountpoint, opts string
}

func parseMountTable(i *image) ([]*procEntry, error) {
	out, err := i.run("cat /proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	entries := []*procEntry{}
	p := &procEntry{}
	for _, line := range strings.Split(out, "\n") {
		if _, err := fmt.Sscanf(line, mountinfoFormat,
			&p.id, &p.parent, &p.major, &p.minor,
			&p.source, &p.mountpoint, &p.opts); err != nil {
			return nil, fmt.Errorf("Scanning '%s' failed: %s", line, err)
		}
		entries = append(entries, p)
	}
	return entries, nil
}

// Looks at /proc/self/mountinfo to determine of the specified
// mountpoint has been mounted
func mounted(i *image, device, mountpoint string) (bool, error) {
	entries, err := parseMountTable(i)
	if err != nil {
		return false, err
	}
	// Search the table for the mountpoint
	for _, entry := range entries {
		if entry.mountpoint == mountpoint || strings.Contains(entry.opts, device) {
			return true, nil
		}
	}
	return false, nil
}
