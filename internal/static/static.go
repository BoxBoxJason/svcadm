package static

import (
	"fmt"
	"os"
)

var (
	SVCADM_HOME = fmt.Sprintf("%s/.svcadm", os.Getenv("HOME"))
)
