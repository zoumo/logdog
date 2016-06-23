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
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	. "github.com/zoumo/go-pythonic"
)

type Handler interface {
	// Handle the specified record, filter and emit
	Handle(*LogRecord)
	// Check if handler should filter the specified record
	Filter(*LogRecord) bool
	// Emit log record to output - e.g. stderr or file
	Emit(*LogRecord)
	// Close output stream, if not return error
	Close() error
}

type NullHandler struct {
	Name string
	ConfigLoader
}

func NewNullHandler() *NullHandler {
	return &NullHandler{}
}

func (self *NullHandler) LoadConfig(config map[string]interface{}) error {
	return nil
}

func (self NullHandler) Handle(*LogRecord) {
	// do nothing
}

func (self NullHandler) Filter(*LogRecord) bool {
	return true
}

func (self NullHandler) Emit(*LogRecord) {
	// do nothing
}

func (self *NullHandler) Close() error {
	return nil
}

// StreamHandler: A handler class which writes logging records,
// appropriately formatted, to a stream.
// Note that this class does not close the stream,
// as os.Stdout or os.Stderr may be used.
type StreamHandler struct {
	Out       io.Writer
	Formatter Formatter
	Name      string
	Level     int
	mu        sync.Mutex
	ConfigLoader
}

func NewStreamHandler() *StreamHandler {
	return &StreamHandler{
		Name:      "",
		Out:       os.Stderr,
		Formatter: TerminalFormatter,
		Level:     NOTHING,
	}
}

func (self *StreamHandler) LoadConfig(c map[string]interface{}) error {
	config, err := DictReflect(c)
	if err != nil {
		return err
	}

	self.Name = config.MustGetString("name", "")

	self.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))

	_formatter := config.MustGetString("formatter", "terminal")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	self.Formatter = formatter

	return nil
}

func (self StreamHandler) Emit(record *LogRecord) {
	msg, err := self.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
	}
	fmt.Fprintln(self.Out, msg)
}

func (self StreamHandler) Filter(record *LogRecord) bool {
	if record.Level < self.Level {
		return true
	}
	return false
}

func (self *StreamHandler) Handle(record *LogRecord) {
	filtered := self.Filter(record)
	if !filtered {
		self.mu.Lock()
		defer self.mu.Unlock()
		self.Emit(record)
	}
}

func (self *StreamHandler) Close() error {
	return nil
}

// File handler
// It is similar to SteamHandler
type FileHandler struct {
	Path string
	Out  *os.File

	Name  string
	Level int

	Formatter Formatter
	mu        sync.Mutex
	ConfigLoader
}

func NewFileHandler() *FileHandler {

	return &FileHandler{
		Name:      "",
		Level:     NOTHING,
		Formatter: DefaultFormatter,
	}
}

func (self *FileHandler) LoadConfig(c map[string]interface{}) error {
	config, err := DictReflect(c)
	if err != nil {
		return nil
	}
	// get name
	self.Name = config.MustGetString("name", "")

	// get path and file
	path := config.MustGetString("filename", "")
	if path == "" {
		return errors.New("Should provide a valid file path")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Errorf("Can not open file %s", path))
	}
	self.Path = path
	self.Out = file

	// get level
	self.Level = GetLevelByName(config.MustGetString("level", "NOTHING"))

	// get formatter
	_formatter := config.MustGetString("formatter", "default")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	self.Formatter = formatter

	return nil
}

func (self *FileHandler) Emit(record *LogRecord) {
	msg, err := self.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
	}
	fmt.Fprintln(self.Out, msg)
}

func (self FileHandler) Filter(record *LogRecord) bool {
	if record.Level < self.Level {
		return true
	}
	return false
}

func (self *FileHandler) Handle(record *LogRecord) {
	if self.Out == nil {
		panic("you should set output file before use this handler")
	}
	filtered := self.Filter(record)
	if !filtered {
		self.mu.Lock()
		defer self.mu.Unlock()
		self.Emit(record)
	}
}
func (self *FileHandler) Close() error {
	if self.Out == nil {
		return nil
	}
	return self.Out.Close()
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
