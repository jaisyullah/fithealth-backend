package logger

import (
	"log"
)

var Log = log.Default()

func Init(level string) {
	// for starter we use default logger; expand to zap/logrus later
	Log.SetPrefix("fithealth: ")
}
