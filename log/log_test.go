package log_test

import (
	"testing"

	"github.com/dawnsgo/dawn/v2/log"
)

func TestLog(t *testing.T) {
	logger := log.NewLogger()

	logger.Debug("welcome to dawn-framework")
	logger.Info("welcome to dawn-framework")
	logger.Warn("welcome to dawn-framework")
	logger.Error("welcome to dawn-framework")
}
