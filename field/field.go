package field

import (
	"bytes"
	"fmt"
)

const Separator = ": "

type Field struct {
	Key   string
	Value string
}

func New(key string, val string) *Field {
	return &Field{
		Key:   key,
		Value: val,
	}
}

func Unmarshal(line []byte) *Field {
	var key, val []byte
	field := &Field{}

	unpack(bytes.Trim(line, Separator), []byte(Separator), &key, &val)

	field.Key = string(key)
	field.Value = string(val)

	return field
}

func (field *Field) Marshal() (line []byte) {
	return []byte(field.String())
}

func (field *Field) String() (str string) {
	var sep string

	if len(field.Value) > 0 {
		sep = Separator
	}

	return fmt.Sprintf("%s%s%s", field.Key, sep, field.Value)
}

func unpack(s, sep []byte, vars ...*[]byte) {
	ss := bytes.SplitN(s, sep, len(vars))

	for i, sss := range ss {
		*vars[i] = sss
	}
}