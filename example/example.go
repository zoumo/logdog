package main

import (
	"github.com/zoumo/logdog"
)

func main() {
	logdog.Info("test")
	logdog.Debug("test")
	logdog.Warn("test")
	logdog.Error("test")
	logdog.Fatal("test")
}
