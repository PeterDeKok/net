package fields

import (
	"bytes"
	"errors"
	"peterdekok.nl/gotools/net/field"
)

const Separator = "\r\n"

type Fields []*field.Field

func New(fs ...*field.Field) *Fields {
	f := &Fields{}

	for _, ff := range fs {
		f.AddField(ff)
	}

	return f
}

func Unmarshal(str []byte) *Fields {
	lines := bytes.Split(bytes.Trim(str, Separator), []byte(Separator))

	fields := Fields{}

	for _, line := range lines {
		fields = append(fields, field.Unmarshal(line))
	}

	return &fields
}

func (f *Fields) Marshal() []byte {
	lines := make([][]byte, len(*f))

	for i, f := range *f {
		lines[i] = f.Marshal()
	}

	return bytes.Join(lines, []byte(Separator))
}

func (f *Fields) String() (str string) {
	return string(f.Marshal())
}

func (f *Fields) AddField(fs ...*field.Field) {
	*f = append(*f, fs...)
}

func (f *Fields) AddFields(ff *Fields) {
	f.AddField(*ff...)
}

func (f *Fields) Find(key string) (value string, err error) {
	for _, f := range *f {
		if f.Key == key {
			return f.Value, nil
		}
	}

	return "", errors.New("field not found")
}
