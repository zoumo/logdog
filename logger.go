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

	"github.com/zoumo/logdog/pkg/pythonic"
)

const (
	DefaultFuncCallDepth = 2
	DEFAULT_FUNC         = 1
)

// Logger All log entries pass through the formatter before logged to Output. The
// included formatters are `TextFormatter` and `JSONFormatter` for which
// TextFormatter is the default. In development (when a TTY is attached) it
// logs with colors, but to a file it wouldn't. You can easily implement your
// own that implements the `Formatter` interface, see the `README` or included
// formatters for examples.
type Logger struct {
	Name     string
	Handlers []Handler
	Level    int
	// funcCallDepth is the number of stack frames to ascend
	funcCallDepth int
	runtimeCaller bool
	ConfigLoader
}

func (lg *Logger) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return nil
	}
	lg.Name = config.MustGetString("name", "")

	lg.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))
	lg.runtimeCaller = config.MustGetBool("enable_runtime_caller", false)

	_handlers := config.MustGetArray("handlers", make([]interface{}, 0))

	for _, h := range _handlers {
		hdlr := GetHandler(h.(string))
		if hdlr == nil {
			panic(fmt.Errorf("can not find handler: %s", h))
		}
		lg.AddHandler(hdlr)
	}

	return nil

}

func (lg *Logger) EnableRuntimeCaller(enable bool) {
	lg.runtimeCaller = enable
}

func (lg *Logger) SetFuncCallDepth(depth int) {
	lg.funcCallDepth = depth
}

func (lg *Logger) SetLevel(level int) {
	lg.Level = level
}

func (lg *Logger) AddHandler(handlers ...Handler) {
	lg.Handlers = append(lg.Handlers, handlers...)
}

func (lg *Logger) log(level int, msg string, args ...interface{}) {
	// 获取runtime的信息
	file := "??"
	line := 0
	funcname := "??"
	if lg.runtimeCaller {
		if _pc, _file, _line, ok := runtime.Caller(lg.funcCallDepth); ok {
			file, line = _file, _line
			if f := runtime.FuncForPC(_pc); f != nil {
				funcname = f.Name() // full func name
			}
		}
	}

	record := NewLogRecord(lg.Name, level, file, funcname, line, msg, args...)
	lg.Handle(record)
}

func (lg *Logger) Handle(record *LogRecord) {
	filtered := lg.Filter(record)
	if !filtered {
		lg.CallHandlers(record)
	}
}

func (lg Logger) Filter(record *LogRecord) bool {
	if record.Level < lg.Level {
		return true
	}
	return false
}

func (lg *Logger) CallHandlers(record *LogRecord) {
	for _, hdlr := range lg.Handlers {
		hdlr.Handle(record)
	}
}

func (lg *Logger) Close() error {
	for _, hdlr := range lg.Handlers {
		err := hdlr.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Close handler failed, [%v]", err)
		}
	}
	return nil
}

func (lg Logger) Logf(level int, msg string, args ...interface{}) {
	lg.log(level, msg, args...)
}

func (lg Logger) Debugf(msg string, args ...interface{}) {
	lg.log(DEBUG, msg, args...)
}
func (lg Logger) Infof(msg string, args ...interface{}) {
	lg.log(INFO, msg, args...)
}
func (lg Logger) Warningf(msg string, args ...interface{}) {
	lg.log(WARN, msg, args...)
}
func (lg Logger) Warnf(msg string, args ...interface{}) {
	lg.log(WARN, msg, args...)
}
func (lg Logger) Errorf(msg string, args ...interface{}) {
	lg.log(ERROR, msg, args...)
}
func (lg Logger) Noticef(msg string, args ...interface{}) {
	lg.log(NOTICE, msg, args...)
}
func (lg Logger) Criticalf(msg string, args ...interface{}) {
	lg.log(CRITICAL, msg, args...)
}
func (lg Logger) Panicf(msg string, args ...interface{}) {
	lg.log(CRITICAL, msg, args...)
	panic("CRITICAL")
}

func (lg Logger) Log(level int, args ...interface{}) {
	lg.log(level, "", args...)
}

func (lg Logger) Debug(args ...interface{}) {
	lg.log(DEBUG, "", args...)
}
func (lg Logger) Info(args ...interface{}) {
	lg.log(INFO, "", args...)
}
func (lg Logger) Warning(args ...interface{}) {
	lg.log(WARN, "", args...)
}
func (lg Logger) Warn(args ...interface{}) {
	lg.log(WARN, "", args...)
}
func (lg Logger) Error(args ...interface{}) {
	lg.log(ERROR, "", args...)
}
func (lg Logger) Notice(args ...interface{}) {
	lg.log(NOTICE, "", args...)
}
func (lg Logger) Critical(args ...interface{}) {
	lg.log(CRITICAL, "", args...)
}
func (lg Logger) Panic(msg string, args ...interface{}) {
	lg.log(CRITICAL, "", args...)
	panic("CRITICAL")
}
