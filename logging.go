// Copyright 2016 Jim Zhang (jim.zoumo@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package logdog

import (
	"fmt"
	"sync"
)

const (
	// NOTHING log level
	NOTHING = 0
	// DEBUG log level
	DEBUG = 1 //0x00000001
	// INFO log level
	INFO = 2 //0x00000010
	// WARN log level
	WARN = 4 //0x00000100
	// WARNING is alias to WARN
	WARNING = 4 //0x00000100
	// ERROR log level
	ERROR = 8 //0x00001000
	// NOTICE log level
	NOTICE = 16 //0x00010000
	// CRITICAL log level
	CRITICAL = 32 //0x00100000
	// ALL log levle
	ALL = 255 //0x11111111
)

var (
	levelNames = map[string]int{
		"NOTHING":  NOTHING,
		"DEBUG":    DEBUG,
		"INFO":     INFO,
		"WARN":     WARN,
		"WARNING":  WARNING,
		"ERROR":    ERROR,
		"NOTICE":   NOTICE,
		"CRITICAL": CRITICAL,
	}
	nameLevels = map[int]string{
		NOTHING:  "NOTHING",
		DEBUG:    "DEBUG",
		INFO:     "INFO",
		WARN:     "WARN",
		ERROR:    "ERROR",
		NOTICE:   "NOTICE",
		CRITICAL: "CRITICAL",
	}
	mu = sync.Mutex{}

	// set default logger
	root = GetLogger("root")
)

// GetLevelName gets level's name
func GetLevelName(level int) string {
	if v, ok := nameLevels[level]; ok {
		return v
	}
	return fmt.Sprintf("level %d", level)

}

// GetLevelByName gets level value by level name
func GetLevelByName(levelname string) int {
	if v, ok := levelNames[levelname]; ok {
		return v
	}

	panic("can not find level by name: " + levelname)
}

// AddLevelName adds new level and level name pair
func AddLevelName(level int, levelName string) {
	mu.Lock()
	defer mu.Unlock()
	levelNames[levelName] = level
	nameLevels[level] = levelName
}

// AddHandler is an alias of root.AddHandler
func AddHandler(handlers ...Handler) {
	root.AddHandler(handlers...)
}

// EnableRuntimeCaller is an alias of root.EnableRuntimeCaller
func EnableRuntimeCaller(enable bool) {
	root.EnableRuntimeCaller(enable)
}

// SetLevel is an alias of root.SetLevel
func SetLevel(level int) {
	root.SetLevel(level)
}

// Close is an alias of root.Close
func Close(formatter Formatter) error {
	return root.Close()
}

// Debugf is an alias of root.Debugf
func Debugf(msg string, args ...interface{}) {
	root.log(DEBUG, msg, args...)
}

// Infof is an alias of root.Infof
func Infof(msg string, args ...interface{}) {
	root.log(INFO, msg, args...)
}

// Warningf is an alias of root.Warningf
func Warningf(msg string, args ...interface{}) {
	root.log(WARN, msg, args...)
}

// Warnf is an alias of root.Warnf
func Warnf(msg string, args ...interface{}) {
	root.log(WARN, msg, args...)
}

// Errorf is an alias of root.Errorf
func Errorf(msg string, args ...interface{}) {
	root.log(ERROR, msg, args...)
}

// Noticef is an alias of root.Noticef
func Noticef(msg string, args ...interface{}) {
	root.log(NOTICE, msg, args...)
}

// Criticalf is an alias of root.Criticalf
func Criticalf(msg string, args ...interface{}) {
	root.log(CRITICAL, msg, args...)
}

// Panicf is an alias of root.Panicf
func Panicf(msg string, args ...interface{}) {
	root.log(CRITICAL, msg, args...)
	panic("CRITICAL")
}

// Debug is an alias of root.Debug
func Debug(args ...interface{}) {
	root.log(DEBUG, "", args...)
}

// Info is an alias of root.Info
func Info(args ...interface{}) {
	root.log(INFO, "", args...)
}

// Warning is an alias of root.Warning
func Warning(args ...interface{}) {
	root.log(WARN, "", args...)
}

// Warn is an alias of root.Warn
func Warn(args ...interface{}) {
	root.log(WARN, "", args...)
}

// Error is an alias of root.Error
func Error(args ...interface{}) {
	root.log(ERROR, "", args...)
}

// Notice is an alias of root.Notice
func Notice(args ...interface{}) {
	root.log(NOTICE, "", args...)
}

// Critical is an alias of root.Critical
func Critical(args ...interface{}) {
	root.log(CRITICAL, "", args...)
}

// Panic an alias of root.Panic
func Panic(msg string, args ...interface{}) {
	root.log(CRITICAL, "", args...)
	panic("CRITICAL")
}

func init() {
	root.AddHandler(NewStreamHandler())
}
