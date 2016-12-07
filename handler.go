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
	"io"
	"os"
	"sync"

	"github.com/zoumo/logdog/pkg/pythonic"
)

var (
	// Discard is an io.ReadWriteCloser on which all Read | Write | Close calls succeed
	// without doing anything.
	Discard = devNull(0)
)

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) Read(p []byte) (n int, err error) {
	return len(p), nil
}

func (devNull) Close() error {
	return nil
}

// Handler specifies how to write a LoadConfig, appropriately formatted, to output.
type Handler interface {
	// Handle the specified record, filter and emit
	Handle(*LogRecord)
	// Filter checks if handler should filter the specified record
	Filter(*LogRecord) bool
	// Emit log record to output - e.g. stderr or file
	Emit(*LogRecord)
	// Close output stream, if not return error
	Close() error
}

// NullHandler is an example handler doing nothing
type NullHandler struct {
	Name string
}

// NewNullHandler returns a NullHandler
func NewNullHandler() *NullHandler {
	return &NullHandler{}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *NullHandler) LoadConfig(config map[string]interface{}) error {
	return nil
}

//Handle the specified record, filter and emit it
func (hdlr *NullHandler) Handle(*LogRecord) {
	// do nothing
}

// Filter checks if handler should filter the specified record
func (hdlr *NullHandler) Filter(*LogRecord) bool {
	return true
}

// Emit log record to output - e.g. stderr or file
func (hdlr *NullHandler) Emit(*LogRecord) {
	// do nothing
}

// Close output stream, if not return error
func (hdlr *NullHandler) Close() error {
	return nil
}

// StreamHandler is a handler which writes logging records,
// appropriately formatted, to a stream.
// Note that this handler does not close the stream,
// as os.Stdout or os.Stderr may be used.
type StreamHandler struct {
	Name      string
	Level     int
	Formatter Formatter
	output    io.Writer
	mu        sync.Mutex
}

// NewStreamHandler returns a new StreamHandler fully initialized
func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		Name:      "",
		output:    os.Stderr,
		Formatter: TerminalFormatter,
		Level:     NothingLevel,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *StreamHandler) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	hdlr.Name = config.MustGetString("name", "")

	hdlr.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))

	_formatter := config.MustGetString("formatter", "terminal")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	hdlr.Formatter = formatter

	return nil
}

// discardOutput replaces handler's Out with Discard, test only
func (hdlr *StreamHandler) discardOutput() *StreamHandler {
	hdlr.output = Discard
	return hdlr
}

// SetOutput changes handler's output
func (hdlr *StreamHandler) SetOutput(out io.Writer) *StreamHandler {
	hdlr.output = out
	return hdlr
}

// SetFormatter changes handler's Formatter
func (hdlr *StreamHandler) SetFormatter(f Formatter) *StreamHandler {
	hdlr.Formatter = f
	return hdlr
}

// SetName changes handler's Formatter
func (hdlr *StreamHandler) SetName(n string) *StreamHandler {
	hdlr.Name = n
	return hdlr
}

// SetLevel changes handler's Level
func (hdlr *StreamHandler) SetLevel(i int) *StreamHandler {
	hdlr.Level = i
	return hdlr
}

// Emit log record to output - e.g. stderr or file
func (hdlr *StreamHandler) Emit(record *LogRecord) {
	msg, err := hdlr.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
	}
	fmt.Fprintln(hdlr.output, msg)
}

// Filter checks if handler should filter the specified record
func (hdlr *StreamHandler) Filter(record *LogRecord) bool {
	if record.Level < hdlr.Level {
		return true
	}
	return false
}

// Handle the specified record, filter and emit it
func (hdlr *StreamHandler) Handle(record *LogRecord) {
	if hdlr.output == nil {
		panic("you should set output file before use this handler")
	}
	filtered := hdlr.Filter(record)
	if !filtered {
		hdlr.mu.Lock()
		defer hdlr.mu.Unlock()
		hdlr.Emit(record)
	}
}

// Close output stream, if not return error
func (hdlr *StreamHandler) Close() error {
	return nil
}

// FileHandler is a handler similar to SteamHandler
// if specified file and it will close the file
type FileHandler struct {
	Name      string
	Level     int
	Formatter Formatter
	output    io.WriteCloser
	Path      string
	mu        sync.Mutex
}

// NewFileHandler returns a new FileHandler fully initialized
func NewFileHandler() *FileHandler {

	return &FileHandler{
		output:    Discard,
		Name:      "",
		Level:     NothingLevel,
		Formatter: DefaultFormatter,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *FileHandler) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}
	// get name
	hdlr.Name = config.MustGetString("name", "")

	// get path and file
	path := config.MustGetString("filename", "")
	hdlr.SetPath(path)

	// get level
	hdlr.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))

	// get formatter
	_formatter := config.MustGetString("formatter", "default")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	hdlr.Formatter = formatter

	return nil
}

// discardOutput replaces handler's Out with Discard, test only
func (hdlr *FileHandler) discardOutput() *FileHandler {
	hdlr.output = Discard
	return hdlr
}

// SetPath opens file located in the path, if not, create it
func (hdlr *FileHandler) SetPath(path string) *FileHandler {
	if path == "" {
		panic("Should provide a valid file path")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("Can not open file %s", path))
	}
	hdlr.Path = path
	hdlr.output = file
	return hdlr
}

// SetFormatter changes handler's Formatter
func (hdlr *FileHandler) SetFormatter(f Formatter) *FileHandler {
	hdlr.Formatter = f
	return hdlr
}

// SetName changes handler's Formatter
func (hdlr *FileHandler) SetName(n string) *FileHandler {
	hdlr.Name = n
	return hdlr
}

// SetLevel changes handler's Level
func (hdlr *FileHandler) SetLevel(i int) *FileHandler {
	hdlr.Level = i
	return hdlr
}

// Emit log record to file
func (hdlr *FileHandler) Emit(record *LogRecord) {
	msg, err := hdlr.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
	}
	fmt.Fprintln(hdlr.output, msg)
}

// Filter checks if handler should filter the specified record
func (hdlr *FileHandler) Filter(record *LogRecord) bool {
	if record.Level < hdlr.Level {
		return true
	}
	return false
}

// Handle the specified record, filter and emit it
func (hdlr *FileHandler) Handle(record *LogRecord) {
	if hdlr.output == nil {
		panic("you should set output file before use this handler")
	}

	filtered := hdlr.Filter(record)
	if !filtered {
		hdlr.mu.Lock()
		defer hdlr.mu.Unlock()
		hdlr.Emit(record)
	}
}

// Close file, if not return error
func (hdlr *FileHandler) Close() error {
	if hdlr.output == nil {
		return nil
	}
	return hdlr.output.Close()
}

func init() {
	RegisterConstructor("NullHandler", func() ConfigLoader {
		return NewNullHandler()
	})
	RegisterConstructor("StreamHandler", func() ConfigLoader {
		return NewStreamHandler()
	})
	RegisterConstructor("FileHandler", func() ConfigLoader {
		return NewFileHandler()
	})

}
