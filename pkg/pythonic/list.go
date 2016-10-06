package pythonic

import "bytes"

// List is a python-like type
type List []interface{}

// NewList returns a new list
func NewList(capacity int) List {
	return make(List, 0, capacity)
}

// Append to the end of list
func (l List) Append(item ...interface{}) List {
	return append(l, item...)
}

// Extend list with another list
func (l List) Extend(list List) List {
	return append(l, list...)
}

// String convert list to a string
func (l List) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("[")
	for _, v := range l {
		buf.WriteString(spprint(v))
		buf.WriteString(", ")
	}
	buf.WriteString("]")
	return buf.String()

}
