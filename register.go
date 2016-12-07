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

import "sync"

var (
	formatters   = NewRegister()
	handlers     = NewRegister()
	constructors = NewRegister()
	loggers      = NewRegister()
)

// Register is a struct binds name and interface such as Constructor
type Register struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewRegister returns a new register
func NewRegister() *Register {
	return &Register{
		data: make(map[string]interface{}),
	}
}

// Register binds name and interface
func (r *Register) Register(name string, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[name]
	if ok {
		panic("Repeated registration key: " + name)
	}
	r.data[name] = v

}

// Get returns an interface registered with given name
func (r *Register) Get(name string) interface{} {
	// need lock ?
	return r.data[name]
}

// Constructor is a function which returns an ConfigLoader
type Constructor func() ConfigLoader

// GetConstructor returns an Constructor registered with given name
// if not, returns nil
func GetConstructor(name string) Constructor {
	v := constructors.Get(name)
	if v == nil {
		return nil
	}
	return v.(Constructor)
}

// RegisterConstructor binds name and Constructor
func RegisterConstructor(name string, c Constructor) {
	constructors.Register(name, c)
}

// RegisterFormatter binds name and Formatter
func RegisterFormatter(name string, formatter Formatter) {
	formatters.Register(name, formatter)
}

// GetFormatter returns an Formatter registered with given name
func GetFormatter(name string) Formatter {
	v := formatters.Get(name)
	if v == nil {
		return nil
	}
	return v.(Formatter)
}

// RegisterHandler binds name and Handler
func RegisterHandler(name string, handler Handler) {
	handlers.Register(name, handler)
}

// GetHandler returns a Handler registered with given name
func GetHandler(name string) Handler {
	v := handlers.Get(name)
	if v == nil {
		return nil
	}
	return v.(Handler)
}

// GetLogger returns an logger by name
// if not, create one and add it to logger register
func GetLogger(name string) *Logger {
	if name == "" {
		name = "root"
	}

	var v interface{}
	v = loggers.Get(name)
	if v != nil {
		return v.(*Logger)
	}

	logger := new(Logger)
	// set name
	logger.Name = name
	// default func call depth is 2
	logger.funcCallDepth = DefaultFuncCallDepth
	// enable analyze runtime caller
	logger.runtimeCaller = true

	// check twice
	// maybe sb. adds logger when this logger is creating
	v = loggers.Get(name)
	if v != nil {
		return v.(*Logger)
	}

	loggers.Register(name, logger)
	return logger
}

// DisableExistingLoggers closes all existing loggers and unregister them
func DisableExistingLoggers() {
	// close all existing logger
	loggers.mu.Lock()
	for _, logger := range loggers.data {
		_logger := logger.(*Logger)
		_logger.Close()
	}
	loggers.data = make(map[string]interface{})
	loggers.mu.Unlock()

	loggers = NewRegister()
	// reset root
	root = GetLogger("root")
	root.AddHandler(NewStreamHandler())
}
