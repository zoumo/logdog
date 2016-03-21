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
package logtar

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Handler interface {
	// Handle the specified record, filter and emit
	Handle(*LogRecord)
	// Check if handler should filter the specified record
	Filter(*LogRecord) bool
	// Emit log record to output - e.g. stderr or file
	Emit(*LogRecord)
	// Close output stream, if not return nil
	Close() error
}

type NullHandler struct {
	Name string
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

func (self NullHandler) Close() error {
	return nil
}

// StreamHandler: A handler class which writes logging records,
// appropriately formatted, to a stream.
// Note that this class does not close the stream,
// as sys.stdout or sys.stderr may be used.
type StreamHandler struct {
	Out       io.Writer
	Formatter Formatter
	Name      string
	Level     int
	mu        sync.Mutex
}

func NewStreamHandler(name string) *StreamHandler {
	return &StreamHandler{
		Name:      name,
		Out:       os.Stderr,
		Formatter: TerminalFormatter,
		Level:     0,
	}
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

func (self StreamHandler) Handle(record *LogRecord) {
	filtered := self.Filter(record)
	if !filtered {
		self.mu.Lock()
		defer self.mu.Unlock()
		self.Emit(record)
	}
}

func (self StreamHandler) Close() error {
	return nil
}
