package log_test

import (
	"testing"

	"github.com/dawnsgo/dawn/log"
)

func TestLog(t *testing.T) {
	logger := log.NewLogger()

	logger.Debug("welcome to dawn-framework")
	logger.Info("welcome to dawn-framework")
	logger.Warn("welcome to dawn-framework")
	logger.Error("welcome to dawn-framework")
}

func TestLogger(t *testing.T) {
	log.SetLogger(log.NewLogger(log.WithLevel(log.LevelDebug)))

	log.Debug("welcome to due-framework")
	log.Info("welcome to due-framework")
	log.Warn("welcome to due-framework")
	log.Error("welcome to due-framework")
}
