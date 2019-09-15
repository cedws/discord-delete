package log

import (
	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init(verbose bool) {
	if verbose {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}
