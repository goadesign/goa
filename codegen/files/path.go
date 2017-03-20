package files

import (
	"fmt"
	"strings"
)

// UniquePath produces a value using format which is not a key of reserved.
// format must be a valid fmt.Printf format using a single verb "%d".
func UniquePath(format string, reserved map[string]bool) string {
	p := strings.Replace(format, "%d", "", -1)
	_, inuse := reserved[p]
	for i := 2; inuse; i++ {
		p = fmt.Sprintf(format, i)
		_, inuse = reserved[p]
	}
	return p
}
