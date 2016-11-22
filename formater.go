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
	"github.com/zoumo/logdog/pkg/pythonic"
	"github.com/zoumo/logdog/pkg/when"
)

// Formatter is an interface which can convert a LogRecord to string
type Formatter interface {
	Format(*LogRecord) (string, error)
}

// FormatTime returns the creation time of the specified LogRecord as formatted text.
func FormatTime(record *LogRecord, datefmt string) string {
	if datefmt == "" {
		datefmt = DefaultDateFmt
	}
	return when.Strftime(&record.Time, datefmt)
}

// TextFormatter is the default formatter used to convert a LogRecord to text.
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
// %(endColor)       reset color
//
type TextFormatter struct {
	Fmt          string
	DateFmt      string
	EnableColors bool
	ConfigLoader
}

const (
	// DefaultFmt is the default log string format value for TextFormatter
	DefaultFmt = "%(color)[%(time)] [%(levelname)] [%(filename):%(lineno)]%(endColor) %(message)"
	// DefaultDateFmt is the default log time string format value for TextFormatter
	DefaultDateFmt = "%Y-%m-%d %H:%M:%S"
	// colors
	blue      = 34
	green     = 32
	yellow    = 33
	red       = 31
	darkGreen = 36
	white     = 37
)

var (
	// LogRecordFieldRegexp is the field regexp
	// for example, I will replace %(name) of real record name
	// TODO support %[(name)][flags][width].[precision]typecode
	LogRecordFieldRegexp = regexp.MustCompile(`\%\(\w+\)`)
	// DefaultFormatter is the default formatter of TextFormatter without color
	DefaultFormatter = TextFormatter{
		Fmt:     DefaultFmt,
		DateFmt: DefaultDateFmt,
	}
	// TerminalFormatter is an TextFormatter with color
	TerminalFormatter = TextFormatter{
		Fmt:          DefaultFmt,
		DateFmt:      DefaultDateFmt,
		EnableColors: true,
	}

	// ColorHash describes colors of different log level
	// you can add new color for your own log level
	ColorHash = map[int]int{
		DEBUG:    blue,
		INFO:     green,
		WARN:     yellow,
		ERROR:    red,
		NOTICE:   darkGreen,
		CRITICAL: red,
	}

	// check if stderr is terminal, sometimes it is redirected to a file
	// isTerminal      = terminal.IsTerminal(syscall.Stderr)
	isTerminal      = term.IsTerminal(uintptr(syscall.Stderr))
	isColorTerminal = isTerminal && (runtime.GOOS != "windows")
)

// colorHash returns color for deferent level, default is white
func colorHash(level int) string {
	// http://blog.csdn.net/acmee/article/details/6613060
	color, ok := ColorHash[level]
	if !ok {
		color = white // white
	}
	return fmt.Sprintf("\033[%dm", color)
}

// NewTextFormatter return a new TextFormatter with default config
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		Fmt:          DefaultFmt,
		DateFmt:      DefaultDateFmt,
		EnableColors: false,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (tf *TextFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	tf.Fmt = config.MustGetString("fmt", DefaultFmt)
	tf.DateFmt = config.MustGetString("datefmt", DefaultDateFmt)
	tf.EnableColors = config.MustGetBool("enable_colors", false)

	return nil

}

// Format converts the specified record to string.
func (tf TextFormatter) Format(record *LogRecord) (string, error) {

	fmtStr := tf.Fmt
	if fmtStr == "" {
		// Don't open color by default
		fmtStr = DefaultFmt
		tf.EnableColors = false
	}
	// 防止需要多次添加颜色, 减少函数调用
	color := ""
	endColor := ""
	if isColorTerminal && tf.EnableColors {
		color = colorHash(record.Level)
		endColor = "\033[0m" // reset color
	}
	fmtStr += tf.formatFields(record)

	// replace %(field) with actual record value
	str := LogRecordFieldRegexp.ReplaceAllStringFunc(fmtStr, func(match string) string {
		// match : %(field)
		field := match[2 : len(match)-1]
		switch field {
		case "name":
			return record.Name
		case "time":
			return FormatTime(record, tf.DateFmt)
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
		case "endColor":
			return endColor
		}
		return match
	})

	return str, nil
}

func (tf TextFormatter) formatFields(record *LogRecord) string {

	b := &bytes.Buffer{}
	b.WriteString("%(color)")
	for k, v := range record.Fields {
		fmt.Fprintf(b, " %s=%+v", k, v)
	}
	b.WriteString("%(endColor)")
	return b.String()
}

// JSONFormatter can convert LogRecord to json text
type JSONFormatter struct {
	Datefmt string
}

// NewJSONFormatter returns a JSONFormatter with default config
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		Datefmt: DefaultDateFmt,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (jf *JSONFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	jf.Datefmt = config.MustGetString("datefmt", DefaultDateFmt)
	return nil
}

// Format converts the specified record to json string.
func (jf JSONFormatter) Format(record *LogRecord) (string, error) {
	fields := make(Fields, len(record.Fields)+4)
	for k, v := range record.Fields {
		fields[k] = v
	}
	// jf.formatFields(fields)
	data := make(map[string]interface{})

	data["time"] = FormatTime(record, jf.Datefmt)
	data["message"] = record.GetMessage()
	data["file"] = record.FileName
	data["line"] = record.Line
	data["level"] = record.LevelName
	if len(fields) > 0 {
		data["_fields"] = fields
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Marashal fields to Json failed, [%v]", err)
	}

	return string(jsonBytes), nil
}

func init() {
	RegisterConstructor("TextFormatter", func() ConfigLoader {
		return NewTextFormatter()
	})
	RegisterConstructor("JsonFormatter", func() ConfigLoader {
		return NewJSONFormatter()
	})

	RegisterFormatter("default", DefaultFormatter)
	RegisterFormatter("terminal", TerminalFormatter)
}
