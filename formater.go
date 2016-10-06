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
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"syscall"

	"github.com/docker/docker/pkg/term"
	. "github.com/zoumo/logdog/pkg/pythonic"
	"github.com/zoumo/logdog/pkg/when"
)

// The Formatter interface to convert a LogRecord to string
type Formatter interface {
	Format(*LogRecord) (string, error)
}

// FormatTime Return the creation time of the specified LogRecord as formatted text.
func FormatTime(record *LogRecord, datefmt string) string {
	if datefmt == "" {
		datefmt = DEFAULT_TIME_FORMAT
	}
	return when.Strftime(&record.Time, datefmt)
}

// Default Formatter instances are used to convert a LogRecord to text.
//
// The Formatter can be initialized with a format string which makes use of
// knowledge of the LogRecord attributes - e.g. the default value mentioned
// above makes use of the fact that the user's message and arguments are pre-
// formatted into a LogRecord's message attribute. Currently, the useful
// attributes in a LogRecord are described by:
//
// %(name)            Name of the logger (logging channel)
// %(levelno)         Numeric logging level for the message (DEBUG, INFO,
//                    WARNING, ERROR, CRITICAL)
// %(levelname)       Text logging level for the message ("DEBUG", "INFO",
//                    "WARNING", "ERROR", "CRITICAL")
// %(pathname)        Full pathname of the source file where the logging
//                    call was issued (if available) or maybe ??
// %(filename)        Filename portion of pathname
// %(lineno)          Source line number where the logging call was issued
//                    (if available)
// %(funcname)        Function name of caller or maybe ??
// %(time)            Textual time when the LogRecord was created
// %(message)         The result of record.getMessage(), computed just as
//                    the record is emitted
// %(color)           print color
// %(end_color)       reset color
//
type TextFormatter struct {
	Fmt          string
	DateFmt      string
	EnableColors bool
	ConfigLoader
}

const (
	DEFAULT_FORMAT      = "%(color)[%(time)] [%(levelname)] [%(filename):%(lineno)]%(end_color) %(message)"
	DEFAULT_TIME_FORMAT = "%Y-%m-%d %H:%M:%S"
	// colors
	blue       = 34
	green      = 32
	yellow     = 33
	red        = 31
	dark_green = 36
	white      = 37
)

var (
	// TODO support %[(name)][flags][width].[precision]typecode
	LogRecordFieldRegexp = regexp.MustCompile(`\%\(\w+\)`)
	// default formatter is an text formatter without color
	DefaultFormatter = TextFormatter{
		Fmt:     DEFAULT_FORMAT,
		DateFmt: DEFAULT_TIME_FORMAT,
	}
	// terminal formatter is an text formatter with color
	TerminalFormatter = TextFormatter{
		Fmt:          DEFAULT_FORMAT,
		DateFmt:      DEFAULT_TIME_FORMAT,
		EnableColors: true,
	}

	// This variable describes colors of different log level
	// you can add new color for your own log level
	ColorHash = map[int]int{
		DEBUG:    blue,
		INFO:     green,
		WARN:     yellow,
		ERROR:    red,
		NOTICE:   dark_green,
		CRITICAL: red,
	}

	// check if stderr is terminal, sometimes it is redirected to a file
	isTerminal      = term.IsTerminal(uintptr(syscall.Stderr))
	isColorTerminal = isTerminal && (runtime.GOOS != "windows")
)

// Return color for deferent level, default is white
func colorHash(level int) string {
	// http://blog.csdn.net/acmee/article/details/6613060
	color, ok := ColorHash[level]
	if !ok {
		color = white // white
	}
	return fmt.Sprintf("\033[%dm", color)
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		Fmt:          DEFAULT_FORMAT,
		DateFmt:      DEFAULT_TIME_FORMAT,
		EnableColors: false,
	}
}

func (self *TextFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := DictReflect(c)
	if err != nil {
		return err
	}

	self.Fmt = config.MustGetString("fmt", DEFAULT_FORMAT)
	self.DateFmt = config.MustGetString("datefmt", DEFAULT_TIME_FORMAT)
	self.EnableColors = config.MustGetBool("enable_colors", false)

	return nil

}

// Format the specified record as text.
func (self TextFormatter) Format(record *LogRecord) (string, error) {

	fmt_str := self.Fmt
	if fmt_str == "" {
		// Don't open color by default
		fmt_str = DEFAULT_FORMAT
		self.EnableColors = false
	}
	// 防止需要多次添加颜色, 减少函数调用
	color := ""
	end_color := ""
	if isColorTerminal && self.EnableColors {
		color = colorHash(record.Level)
		end_color = "\033[0m" // reset color
	}
	fmt_str += self.formatFields(record)

	// replace %(field) with actual record value
	str := LogRecordFieldRegexp.ReplaceAllStringFunc(fmt_str, func(match string) string {
		// match : %(field)
		field := match[2 : len(match)-1]
		switch field {
		case "name":
			return record.Name
		case "time":
			return FormatTime(record, self.DateFmt)
		case "levelno":
			return fmt.Sprintf("%d", record.Level)
		case "levelname":
			return record.LevelName
		case "pathname":
			return record.PathName
		case "filename":
			return record.FileName
		case "funcname":
			return record.ShortFuncName
		case "lineno":
			return fmt.Sprintf("%d", record.Line)
		case "message":
			return record.GetMessage()
		case "color":
			return color
		case "end_color":
			return end_color
		}
		return match
	})

	return str, nil
}

func (self TextFormatter) formatFields(record *LogRecord) string {

	b := &bytes.Buffer{}
	b.WriteString("%(color)")
	for k, v := range record.Fields {
		fmt.Fprintf(b, " %s=%+v", k, v)
	}
	b.WriteString("%(end_color)")
	return b.String()
}

// JsonFormatter
type JsonFormatter struct {
	Datefmt string
}

func NewJsonFormatter() *JsonFormatter {
	return &JsonFormatter{
		Datefmt: DEFAULT_TIME_FORMAT,
	}
}

func (self *JsonFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := DictReflect(c)
	if err != nil {
		return err
	}

	self.Datefmt = config.MustGetString("datefmt", DEFAULT_TIME_FORMAT)
	return nil
}

func (self JsonFormatter) Format(record *LogRecord) (string, error) {
	fields := make(Fields, len(record.Fields)+4)
	for k, v := range record.Fields {
		fields[k] = v
	}
	// self.formatFields(fields)
	data := make(map[string]interface{})

	data["time"] = FormatTime(record, self.Datefmt)
	data["message"] = record.GetMessage()
	data["file"] = record.FileName
	data["line"] = record.Line
	data["level"] = record.LevelName
	if len(fields) > 0 {
		data["_fields"] = fields
	}

	json_bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Marashal fields to Json failed, [%v]", err)
	}

	return string(json_bytes), nil
}

func init() {
	RegisterConstructor("TextFormatter", func() ConfigLoader {
		return NewTextFormatter()
	})
	RegisterConstructor("JsonFormatter", func() ConfigLoader {
		return NewJsonFormatter()
	})

	RegisterFormatter("default", DefaultFormatter)
	RegisterFormatter("terminal", TerminalFormatter)
}
