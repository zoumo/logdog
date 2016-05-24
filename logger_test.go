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

func TestLogger(t *testing.T) {
	// record := NewLogRecord(name, DEBUG, pathname, fun, line, "%s", "debug", fields)
	logger := GetLogger("test")
	assert.Equal(t, "test", logger.Name)

	logger.AddHandler(NewStreamHandler())
	assert.Len(t, logger.Handlers, 1)

	logger.SetLevel(INFO)
	assert.Equal(t, logger.Level, INFO)

	logger.SetFuncCallDepth(3)
	assert.Equal(t, 3, logger.funcCallDepth)

	logger.EnableRuntimeCaller(false)
	logger.Debug("test debug") // filtered
	logger.Info("test info")
	logger.Critical("test critical")

}

func TestJsonLogger(t *testing.T) {
	logger := GetLogger("json")
	formatter := JsonFormatter{}
	handler := NewStreamHandler()
	handler.Formatter = formatter
	logger.AddHandler(handler)
	logger.Info("this is json formatter1")
	logger.Notice("this is json formatter2", fields)

}
