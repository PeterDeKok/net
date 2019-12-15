package message

import (
	"bytes"
	"fmt"
	"peterdekok.nl/gotools/net/field"
	"peterdekok.nl/gotools/net/fields"
)

const Separator = fields.Separator + fields.Separator

type Message struct {
	fields *fields.Fields
	raw []byte
}

func New(fs ...*field.Field) *Message {
	return &Message{
		fields: fields.New(fs...),

		// There is technically no 'raw' message,
		// but we could decide to generate it...
		raw:    []byte(nil),
	}
}

func Unmarshal(str []byte) *Message {
	s := bytes.Trim(str, Separator)

	return &Message{
		fields: fields.Unmarshal(s),
		raw: append(str, []byte(nil)...),
	}
}

func (msg *Message) Marshal() (str []byte) {
	return append(msg.fields.Marshal(), []byte(Separator)...)
}

func (msg *Message) String() (str string) {
	return fmt.Sprintf("%s%s", msg.fields.String(), Separator)
}

func (msg *Message) GetFields() *fields.Fields {
	return msg.fields
}

func (msg *Message) GetField(key string) (val string, err error) {
	return msg.fields.Find(key)
}

func (msg *Message) GetRaw() []byte {
	return msg.raw
}

func (msg *Message) AddFields(f *fields.Fields) {
	msg.fields.AddFields(f)
}

func (msg *Message) AddField(f ...*field.Field) {
	msg.fields.AddField(f...)
}
