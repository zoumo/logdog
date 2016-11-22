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
	// DefaultFuncCallDepth is 2 because you should ascend 2 frames
	// to get true caller function by default
	DefaultFuncCallDepth = 2
)

// Logger entries pass through the formatter before logged to Output. The
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
	// you should change it if you implement your own log function
	funcCallDepth int
	runtimeCaller bool
	ConfigLoader
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
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

// EnableRuntimeCaller enables logger to get caller so that logger
// can get caller's filename and lineno
func (lg *Logger) EnableRuntimeCaller(enable bool) {
	lg.runtimeCaller = enable
}

// SetFuncCallDepth changes the number of stack frames to ascend
// you should change it if you implement your own log api function, e.g. Notify
func (lg *Logger) SetFuncCallDepth(depth int) {
	lg.funcCallDepth = depth
}

// SetLevel defines the filter level
func (lg *Logger) SetLevel(level int) {
	lg.Level = level
}

// AddHandler adds handler to logger
func (lg *Logger) AddHandler(handlers ...Handler) {
	lg.Handlers = append(lg.Handlers, handlers...)
}

// log is the true logging function
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

// Handle handles the LogRecord, call all halders
func (lg *Logger) Handle(record *LogRecord) {
	filtered := lg.Filter(record)
	if !filtered {
		lg.CallHandlers(record)
	}
}

// Filter checks if logger should filter the specified record
func (lg Logger) Filter(record *LogRecord) bool {
	if record.Level < lg.Level {
		return true
	}
	return false
}

// CallHandlers call all handler registered in logger
func (lg *Logger) CallHandlers(record *LogRecord) {
	for _, hdlr := range lg.Handlers {
		hdlr.Handle(record)
	}
}

// Close closes output stream
func (lg *Logger) Close() error {
	for _, hdlr := range lg.Handlers {
		err := hdlr.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Close handler failed, [%v]", err)
		}
	}
	return nil
}

// Logf emits log with specified level and format string
func (lg Logger) Logf(level int, msg string, args ...interface{}) {
	lg.log(level, msg, args...)
}

// Debugf emits log with DEBUG level and format string
func (lg Logger) Debugf(msg string, args ...interface{}) {
	lg.log(DEBUG, msg, args...)
}

// Infof emits log with INFO level and format string
func (lg Logger) Infof(msg string, args ...interface{}) {
	lg.log(INFO, msg, args...)
}

// Warningf emits log with WARN level and format string
func (lg Logger) Warningf(msg string, args ...interface{}) {
	lg.log(WARN, msg, args...)
}

// Warnf emits log with WARN level and format string
func (lg Logger) Warnf(msg string, args ...interface{}) {
	lg.log(WARN, msg, args...)
}

// Errorf emits log with ERROR level and format string
func (lg Logger) Errorf(msg string, args ...interface{}) {
	lg.log(ERROR, msg, args...)
}

// Noticef emits log with NOTICE level and format string
func (lg Logger) Noticef(msg string, args ...interface{}) {
	lg.log(NOTICE, msg, args...)
}

// Criticalf emits log with CRITICAL level and format string
func (lg Logger) Criticalf(msg string, args ...interface{}) {
	lg.log(CRITICAL, msg, args...)
}

// Panicf emits log with CRITICAL level and format string
// and panic it
func (lg Logger) Panicf(msg string, args ...interface{}) {
	lg.log(CRITICAL, msg, args...)
	panic("CRITICAL")
}

// Log emits log message
func (lg Logger) Log(level int, args ...interface{}) {
	lg.log(level, "", args...)
}

// Debug emits log message with DEBUG level
func (lg Logger) Debug(args ...interface{}) {
	lg.log(DEBUG, "", args...)
}

//Info emits log message with INFO level
func (lg Logger) Info(args ...interface{}) {
	lg.log(INFO, "", args...)
}

// Warning emits log message with WARN level
func (lg Logger) Warning(args ...interface{}) {
	lg.log(WARN, "", args...)
}

// Warn emits log message with WARN level
func (lg Logger) Warn(args ...interface{}) {
	lg.log(WARN, "", args...)
}

// Error emits log message with ERROR level
func (lg Logger) Error(args ...interface{}) {
	lg.log(ERROR, "", args...)
}

// Notice emits log message with NOTICE level
func (lg Logger) Notice(args ...interface{}) {
	lg.log(NOTICE, "", args...)
}

// Critical emits log message with CRITICAL level
func (lg Logger) Critical(args ...interface{}) {
	lg.log(CRITICAL, "", args...)
}

// Panic emits log message with CRITICAL level
// and panic it
func (lg Logger) Panic(msg string, args ...interface{}) {
	lg.log(CRITICAL, "", args...)
	panic("CRITICAL")
}
