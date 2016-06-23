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
	NOTHING  = 0
	DEBUG    = 1   //0x00000001
	INFO     = 2   //0x00000010
	WARN     = 4   //0x00000100
	WARNING  = 4   //0x00000100
	ERROR    = 8   //0x00001000
	NOTICE   = 16  //0x00010000
	CRITICAL = 32  //0x00100000
	ALL      = 255 //0x11111111
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
	mu sync.Mutex = sync.Mutex{}

	// set default logger
	root *Logger = GetLogger("root")
)

// Get level name by level
func GetLevelName(level int) string {
	v, ok := nameLevels[level]
	if ok {
		return v
	} else {
		return fmt.Sprintf("level %d", level)
	}
}

func GetLevelByName(levelname string) int {
	v, ok := levelNames[levelname]
	if ok {
		return v
	} else {
		panic("can not find level by name: " + levelname)
	}
}

// Add new level and level name
func AddLevelName(level int, levelName string) {
	mu.Lock()
	defer mu.Unlock()
	levelNames[levelName] = level
	nameLevels[level] = levelName
}

// This function is an alias of root.AddHandler
func AddHandler(handlers ...Handler) {
	root.AddHandler(handlers...)
}

// This function is an alias of root.EnableRuntimeCaller
func EnableRuntimeCaller(enable bool) {
	root.EnableRuntimeCaller(enable)
}

// This function is an alias of root.SetLevel
func SetLevel(level int) {
	root.SetLevel(level)
}

// This function is an alias of root.Close
func Close(formatter Formatter) error {
	return root.Close()
}

func Debugf(msg string, args ...interface{}) {
	root.log(DEBUG, msg, args...)
}
func Infof(msg string, args ...interface{}) {
	root.log(INFO, msg, args...)
}
func Warningf(msg string, args ...interface{}) {
	root.log(WARN, msg, args...)
}
func Warnf(msg string, args ...interface{}) {
	root.log(WARN, msg, args...)
}
func Errorf(msg string, args ...interface{}) {
	root.log(ERROR, msg, args...)
}
func Noticef(msg string, args ...interface{}) {
	root.log(NOTICE, msg, args...)
}
func Criticalf(msg string, args ...interface{}) {
	root.log(CRITICAL, msg, args...)
}
func Panicf(msg string, args ...interface{}) {
	root.log(CRITICAL, msg, args...)
	panic("CRITICAL")
}

func Debug(args ...interface{}) {
	root.log(DEBUG, "", args...)
}
func Info(args ...interface{}) {
	root.log(INFO, "", args...)
}
func Warning(args ...interface{}) {
	root.log(WARN, "", args...)
}
func Warn(args ...interface{}) {
	root.log(WARN, "", args...)
}
func Error(args ...interface{}) {
	root.log(ERROR, "", args...)
}
func Notice(args ...interface{}) {
	root.log(NOTICE, "", args...)
}
func Critical(args ...interface{}) {
	root.log(CRITICAL, "", args...)
}
func Panic(msg string, args ...interface{}) {
	root.log(CRITICAL, "", args...)
	panic("CRITICAL")
}

func init() {
	root.AddHandler(NewStreamHandler())
}
