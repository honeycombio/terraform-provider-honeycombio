package log

import (
	"os"
	"strings"
)

const envLog = "TF_LOG"

func IsDebugOrHigher() bool {
	level := strings.ToUpper(os.Getenv(envLog))
	return level == "DEBUG" || level == "TRACE"
}
