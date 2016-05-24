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
	"path"
	"strings"
	"time"
)

type Fields map[string]interface{}

func (self Fields) String() string {
	return fmt.Sprintf("%#v", (map[string]interface{})(self))
}

type LogRecord struct {
	Name          string
	Level         int
	LevelName     string
	PathName      string
	FileName      string
	FuncName      string
	ShortFuncName string
	Line          int
	Time          time.Time
	// msg could be ""
	Msg  string
	Args []interface{}
	// extract fields from args
	Fields Fields
}

func NewLogRecord(name string, level int, pathname string, funcname string, line int, msg string, args ...interface{}) *LogRecord {
	record := LogRecord{
		Name:     name,
		Level:    level,
		PathName: pathname,
		Line:     line,
		Msg:      msg,
		Args:     args,
		Time:     time.Now(),
	}
	// level name
	record.LevelName = GetLevelName(level)

	// file name
	_, filename := path.Split(pathname)
	record.FileName = filename

	// func name
	i := strings.LastIndex(funcname, "/")
	record.FuncName = funcname[i+1:]
	j := strings.LastIndex(funcname[i+1:], ".")
	record.ShortFuncName = record.FuncName[j+1:]

	// split args and fields
	record.ExtractFieldsFromArgs()

	return &record
}

// Format record message by msg and args
// if msg == "" {
//     msg = fmt.Sprint(self.Args...)
// } else {
//     msg = fmt.Sprintf(self.Msg, self.Args...)
// }
func (self LogRecord) GetMessage() string {
	msg := self.Msg
	if msg == "" {
		msg = fmt.Sprint(self.Args...)
	} else {
		msg = fmt.Sprintf(self.Msg, self.Args...)
	}
	return msg
}

// Extract fields (Fields) from args
// Fields should be the last element in args
func (self *LogRecord) ExtractFieldsFromArgs() {
	args_len := len(self.Args)
	if args_len == 0 {
		return
	}

	if fields, ok := self.Args[args_len-1].(Fields); ok {
		self.Args = self.Args[:args_len-1]
		self.Fields = fields
	}

}
