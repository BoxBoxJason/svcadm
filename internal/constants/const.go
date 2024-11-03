package constants

import (
	"fmt"
	"os"
)

var (
	SVCADM_HOME   = fmt.Sprintf("%s/.svcadm", os.Getenv("HOME"))
	LOG_DIR       = fmt.Sprintf("%s/logs", SVCADM_HOME)
	SVCADM_LABELS = map[string]string{
		"managed_by":     "svcadm",
		"svcadm_version": SVCADM_VERSION,
	}
	SVCADM_VERSION = "latest-dev"
)
