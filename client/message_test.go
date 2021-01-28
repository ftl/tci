package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMessage(t *testing.T) {
	tt := []struct {
		value    string
		expected Message
		invalid  bool
	}{
		{value: "", invalid: true},
		{value: "start", invalid: true},
		{value: "dds0,123;", invalid: true},
		{value: "start;", expected: NewMessage("start")},
		{value: "dds:0,123;", expected: NewMessage("dds", 0, 123)},
		{value: "dds:0,SomeName;", expected: NewMessage("dds", 0, "SomeName")},
		{value: "if:0,-1200;", expected: NewMessage("if", 0, -1200)},
		{value: "rit_enable:0,true;", expected: NewMessage("rit_enable", 0, true)},
		{value: "tx_power:13.5;", expected: NewMessage("tx_power", 13.5)},
	}
	for _, tc := range tt {
		t.Run(tc.value, func(t *testing.T) {
			actual, err := ParseMessage(tc.value)
			if tc.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestIsReplyTo(t *testing.T) {
	assert.True(t, NewMessage("name", 1, 2).IsReplyTo(NewMessage("name", 1)))
	assert.False(t, NewMessage("name", 0, 2).IsReplyTo(NewMessage("name", 1)))
}

func TestToFloat(t *testing.T) {
	f, err := NewMessage("tx_power", 13.5).ToFloat(0)
	assert.NoError(t, err)
	assert.Equal(t, 13.5, f)
}
