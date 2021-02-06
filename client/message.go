package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var messageExp = regexp.MustCompile(`(?P<name>[A-Za-z_]+)(:(?P<args>[A-Za-z0-9-.]+(,[A-Za-z0-9-.]+)*))?;`)

// ParseMessage parses the given string as a TCI message.
func ParseTextMessage(s string) (Message, error) {
	matches := messageExp.FindStringSubmatch(s)
	if len(matches) == 0 {
		return Message{}, fmt.Errorf("invalid message format: %s", s)
	}

	nameIndex := messageExp.SubexpIndex("name")
	if nameIndex == -1 {
		return Message{}, fmt.Errorf("invalid message format, name not found: %s", s)
	}
	name := strings.ToLower(strings.TrimSpace(matches[nameIndex]))

	argsIndex := messageExp.SubexpIndex("args")
	var args []string
	if argsIndex == -1 || matches[argsIndex] == "" {
		args = []string{}
	} else {
		args = strings.Split(matches[argsIndex], ",")
	}

	return Message{name: name, args: args}, nil
}

// NewMessage returns a new message with the given name and the given arguments.
func NewMessage(name string, args ...interface{}) Message {
	result := Message{
		name: strings.ToLower(strings.TrimSpace(name)),
		args: make([]string, len(args)),
	}
	for i, arg := range args {
		result.args[i] = strings.TrimSpace(fmt.Sprintf("%v", arg))
	}
	return result
}

// Message represents a message that is exchanged between the TCI server and a client.
type Message struct {
	name string
	args []string
}

func (m Message) String() string {
	if len(m.args) == 0 {
		return fmt.Sprintf("%s;", m.name)
	}
	return fmt.Sprintf("%s:%s;", m.name, strings.Join(m.args, ","))
}

// IsReplyTo indicates if this message is a reply to the given message.
func (m Message) IsReplyTo(o Message) bool {
	otherString := o.String()
	prefix := otherString[0 : len(otherString)-1]
	return strings.HasPrefix(m.String(), prefix)
}

// Name of the message
func (m Message) Name() string {
	return m.name
}

// Args of the message
func (m Message) Args() []string {
	return m.args
}

func (m Message) arg(i int) (string, error) {
	if len(m.args) < i+1 {
		return "", fmt.Errorf("invalid argument index %d: %q", i, m)
	}
	return m.args[i], nil
}

// ToString returns the argument with the given index as string.
func (m Message) ToString(i int) (string, error) {
	return m.arg(i)
}

// ToInt returns the argument with the given index as integer.
func (m Message) ToInt(i int) (int, error) {
	arg, err := m.arg(i)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(arg)
}

// ToBool returns the argument with the given index as boolean.
func (m Message) ToBool(i int) (bool, error) {
	arg, err := m.arg(i)
	if err != nil {
		return false, err
	}
	return arg == "true" || arg == "1", nil
}

// ToFloat returns the argument with the given index as float.
func (m Message) ToFloat(i int) (float64, error) {
	arg, err := m.arg(i)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(arg, 64)
}

func ParseBinaryMessage(b []byte) (BinaryMessage, error) {
	buf := bytes.NewReader(b)
	var msg encodedBinaryMessage
	err := binary.Read(buf, binary.LittleEndian, &msg)
	if err != nil {
		return BinaryMessage{}, fmt.Errorf("cannot read binary message header: %v", err)
	}

	var data []float32
	if BinaryMessageType(msg.Type) != TXChronoMessage && msg.DataLength > 0 {
		data = make([]float32, msg.DataLength)
		err = binary.Read(buf, binary.LittleEndian, &data)
		if err != nil {
			return BinaryMessage{}, fmt.Errorf("cannot read binary message data: %d %d %v", msg.Type, msg.DataLength, err)
		}
	}

	result := BinaryMessage{
		TRX:        int(msg.TRX),
		SampleRate: int(msg.SampleRate),
		Format:     int(msg.Format),
		Codec:      int(msg.Codec),
		CRC:        msg.CRC,
		DataLength: msg.DataLength,
		Type:       BinaryMessageType(msg.Type),
		Data:       data,
	}

	return result, nil
}

type encodedBinaryMessage struct {
	TRX        uint32
	SampleRate uint32
	Format     uint32
	Codec      uint32
	CRC        uint32
	DataLength uint32
	Type       uint32
	Reserved   [9]uint32
}

type BinaryMessage struct {
	TRX        int
	SampleRate int
	Format     int
	Codec      int
	CRC        uint32
	DataLength uint32
	Type       BinaryMessageType
	Data       []float32
}

type BinaryMessageType uint32

const (
	IQStreamMessage BinaryMessageType = iota
	RXAudioStreamMessage
	TXAudioStreamMessage
	TXChronoMessage
)
