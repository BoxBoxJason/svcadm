package constants

import (
	"os"
	"path"
)

var (
	SVCADM_HOME   = path.Join(os.Getenv("HOME"), ".svcadm")
	LOG_DIR       = path.Join(SVCADM_HOME, "logs")
	SVCADM_LABELS = map[string]string{
		"managed_by":     "svcadm",
		"svcadm_version": SVCADM_VERSION,
	}
	SVCADM_VERSION = "latest-dev"
)
