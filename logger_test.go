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
	logger := GetLogger("test")
	logger.AddHandler(NewStreamHandler().discardOutput()).
		SetLevel(INFO).
		SetFuncCallDepth(3)

	assert.Equal(t, "test", logger.Name)
	assert.Len(t, logger.Handlers, 1)
	assert.Equal(t, logger.Level, INFO)
	assert.Equal(t, 3, logger.funcCallDepth)

	// echo
	logger.SetFuncCallDepth(2).EnableRuntimeCaller(true)
	logger.Debug("test debug") // filtered

	logger.SetLevel(NOTHING)

	logger.Debug("who is your daddy", Fields{"who": "jim"})
	logger.Info("logdog is useful", Fields{"agree": "yes"})
	logger.Warn("warning warning", Fields{"x": "man"})
	logger.Notice("this notice is impotant", Fields{"x": "man"})
	logger.Error("error error..", Fields{"x": "man"})
	logger.Critical("I have no idea !", Fields{"x": "man"})

}

func TestJsonLogger(t *testing.T) {
	logger := GetLogger("json").AddHandler(
		NewStreamHandler().
			SetFormatter(NewJSONFormatter()).
			discardOutput(),
	)

	logger.Info("this is json formatter1")
	logger.Notice("this is json formatter2", fields)

}

func TestLoggerInterface(t *testing.T) {
	assert.Implements(t, (*ConfigLoader)(nil), NewLogger())
}
