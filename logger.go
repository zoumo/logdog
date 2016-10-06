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
	"os"
	"runtime"

	. "github.com/zoumo/logdog/pkg/pythonic"

)

const (
	DEFAULT_FUNC_CALL_DEPTH = 2
)

// All log entries pass through the formatter before logged to Out. The
// included formatters are `TextFormatter` and `JSONFormatter` for which
// TextFormatter is the default. In development (when a TTY is attached) it
// logs with colors, but to a file it wouldn't. You can easily implement your
// own that implements the `Formatter` interface, see the `README` or included
// formatters for examples.
type Logger struct {
	Name          string
	Handlers      []Handler
	Level         int
	funcCallDepth int
	runtimeCaller bool
	ConfigLoader
}

func (self *Logger) LoadConfig(c map[string]interface{}) error {
	config, err := DictReflect(c)
	if err != nil {
		return nil
	}
	self.Name = config.MustGetString("name", "")

	self.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))
	self.runtimeCaller = config.MustGetBool("enable_runtime_caller", false)

	_handlers := config.MustGetArray("handlers", make([]interface{}, 0))

	for _, h := range _handlers {
		hdlr := GetHandler(h.(string))
		if hdlr == nil {
			panic(fmt.Errorf("can not find handler: %s", h))
		}
		self.AddHandler(hdlr)
	}

	return nil

}

func (self *Logger) EnableRuntimeCaller(enable bool) {
	self.runtimeCaller = enable
}

func (self *Logger) SetFuncCallDepth(depth int) {
	self.funcCallDepth = depth
}

func (self *Logger) SetLevel(level int) {
	self.Level = level
}

func (self *Logger) AddHandler(handlers ...Handler) {
	self.Handlers = append(self.Handlers, handlers...)
}

func (self Logger) log(level int, msg string, args ...interface{}) {
	// 获取runtime的信息
	file := "??"
	line := 0
	funcname := "??"
	if self.runtimeCaller {
		if _pc, _file, _line, ok := runtime.Caller(self.funcCallDepth); ok {
			file, line = _file, _line
			if f := runtime.FuncForPC(_pc); f != nil {
				funcname = f.Name() // full func name
			}
		}
	}

	record := NewLogRecord(self.Name, level, file, funcname, line, msg, args...)
	self.Handle(record)
}

func (self Logger) Handle(record *LogRecord) {
	filtered := self.Filter(record)
	if !filtered {
		self.CallHandlers(record)
	}
}

func (self Logger) Filter(record *LogRecord) bool {
	if record.Level < self.Level {
		return true
	}
	return false
}

func (self Logger) CallHandlers(record *LogRecord) {
	for _, hdlr := range self.Handlers {
		hdlr.Handle(record)
	}
}

func (self Logger) Close() error {
	for _, hdlr := range self.Handlers {
		err := hdlr.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Close handler failed, [%v]", err)
		}
	}
	return nil
}

func (self Logger) Logf(level int, msg string, args ...interface{}) {
	self.log(level, msg, args...)
}

func (self Logger) Debugf(msg string, args ...interface{}) {
	self.log(DEBUG, msg, args...)
}
func (self Logger) Infof(msg string, args ...interface{}) {
	self.log(INFO, msg, args...)
}
func (self Logger) Warningf(msg string, args ...interface{}) {
	self.log(WARN, msg, args...)
}
func (self Logger) Warnf(msg string, args ...interface{}) {
	self.log(WARN, msg, args...)
}
func (self Logger) Errorf(msg string, args ...interface{}) {
	self.log(ERROR, msg, args...)
}
func (self Logger) Noticef(msg string, args ...interface{}) {
	self.log(NOTICE, msg, args...)
}
func (self Logger) Criticalf(msg string, args ...interface{}) {
	self.log(CRITICAL, msg, args...)
}
func (self Logger) Panicf(msg string, args ...interface{}) {
	self.log(CRITICAL, msg, args...)
	panic("CRITICAL")
}

func (self Logger) Log(level int, args ...interface{}) {
	self.log(level, "", args...)
}

func (self Logger) Debug(args ...interface{}) {
	self.log(DEBUG, "", args...)
}
func (self Logger) Info(args ...interface{}) {
	self.log(INFO, "", args...)
}
func (self Logger) Warning(args ...interface{}) {
	self.log(WARN, "", args...)
}
func (self Logger) Warn(args ...interface{}) {
	self.log(WARN, "", args...)
}
func (self Logger) Error(args ...interface{}) {
	self.log(ERROR, "", args...)
}
func (self Logger) Notice(args ...interface{}) {
	self.log(NOTICE, "", args...)
}
func (self Logger) Critical(args ...interface{}) {
	self.log(CRITICAL, "", args...)
}
func (self Logger) Panic(msg string, args ...interface{}) {
	self.log(CRITICAL, "", args...)
	panic("CRITICAL")
}
