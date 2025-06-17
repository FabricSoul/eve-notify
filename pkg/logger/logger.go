package logger

import (
	"log"

	"go.uber.org/zap"
)

// Sugar is a global instance of the zap sugared logger.
var Sugar *zap.SugaredLogger

// Init initializes the global logger and returns a cleanup function.
// The cleanup function should be deferred in the main function.
func Init() func() {
	// NewProduction builds a sensible production logger that writes InfoLevel and
	// above logs to standard error.
	coreLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	// Assign the sugared logger to our global variable.
	Sugar = coreLogger.Sugar()

	// Return the Sync function to be deferred by the caller.
	return func() {
		_ = coreLogger.Sync()
	}
}
