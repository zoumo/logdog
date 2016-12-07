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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamHandler(t *testing.T) {
	record := NewLogRecord(name, DEBUG, pathname, fun, line, "%s", "debug", fields)
	record2 := NewLogRecord(name, INFO, pathname, fun, line, "%s", "success", fields)
	handler := NewStreamHandler()
	// handler.Level = INFO
	err := handler.LoadConfig(Config{
		"level": "INFO",
	})
	assert.Nil(t, err)
	assert.Equal(t, handler.Level, INFO)
	assert.Equal(t, handler.Formatter, TerminalFormatter)
	assert.True(t, handler.Filter(record))
	assert.False(t, handler.Filter(record2))
	assert.Nil(t, handler.Close())

}

func TestFileHandler(t *testing.T) {
	record := NewLogRecord(name, DEBUG, pathname, fun, line, "%s", "debug", fields)
	record2 := NewLogRecord(name, INFO, pathname, fun, line, "%s", "success", fields)
	handler := NewFileHandler()

	err := handler.LoadConfig(Config{
		"level":     "INFO",
		"filename":  "./test",
		"formatter": "terminal",
	})
	assert.Nil(t, err)
	assert.Equal(t, handler.Path, "./test")
	assert.Equal(t, handler.Formatter, TerminalFormatter)
	assert.Equal(t, handler.Level, INFO)

	assert.True(t, handler.Filter(record))
	assert.False(t, handler.Filter(record2))
	assert.Nil(t, handler.Close())
}

func TestHandlerInterface(t *testing.T) {
	assert.Implements(t, (*Handler)(nil), NewStreamHandler())
	assert.Implements(t, (*ConfigLoader)(nil), NewStreamHandler())
	assert.Implements(t, (*Handler)(nil), NewFileHandler())
	assert.Implements(t, (*ConfigLoader)(nil), NewFileHandler())
}
