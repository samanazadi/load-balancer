package logging

import "log"

var Logger *log.Logger

func Init() {
	Logger = log.Default()
}
