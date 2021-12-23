package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTextMessage(t *testing.T) {
	tt := []struct {
		value    string
		expected Message
		invalid  bool
	}{
		{value: "", invalid: true},
		{value: "start", invalid: true},
		{value: "dds0,123;", invalid: true},
		{value: "start;", expected: NewCommandMessage("start")},
		{value: "dds:0,123;", expected: NewCommandMessage("dds", 0, 123)},
		{value: "dds:0,SomeName;", expected: NewCommandMessage("dds", 0, "SomeName")},
		{value: "if:0,-1200;", expected: NewCommandMessage("if", 0, -1200)},
		{value: "rit_enable:0,true;", expected: NewCommandMessage("rit_enable", 0, true)},
		{value: "tx_power:13.5;", expected: NewCommandMessage("tx_power", 13.5)},
	}
	for _, tc := range tt {
		t.Run(tc.value, func(t *testing.T) {
			actual, err := ParseTextMessage(tc.value)
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
	assert.True(t, NewCommandMessage("name", 1, 2).IsReplyTo(NewRequestMessage("name", 1)))
	assert.False(t, NewCommandMessage("name", 0, 2).IsReplyTo(NewRequestMessage("name", 1)))
}

func TestToFloat(t *testing.T) {
	f, err := NewCommandMessage("tx_power", 13.5).ToFloat(0)
	assert.NoError(t, err)
	assert.Equal(t, 13.5, f)
}
