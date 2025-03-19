package backendruntime

import "os"

var (
	AvailableBackends []*Backend
	RunningBackends   map[uint]*Runtime
	TempDir           string
	shouldLog         bool
)

func init() {
	RunningBackends = make(map[uint]*Runtime)
	shouldLog = os.Getenv("HERMES_DEVELOPMENT_MODE") != "" || os.Getenv("HERMES_BACKEND_LOGGING_ENABLED") != "" || os.Getenv("HERMES_LOG_LEVEL") == "debug"
}
