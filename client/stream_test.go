package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampleBuffer_Length(t *testing.T) {
	tt := []struct {
		desc     string
		capacity int
		read     int
		write    int
		expected int
	}{
		{"empty", 10, 0, 0, 0},
		{"1", 10, 0, 1, 1},
		{"5 over edge", 10, 8, 2, 5},
		{"10", 10, 1, 0, 10},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			b := newSampleBuffer(tc.capacity)
			b.read = tc.read
			b.write = tc.write
			assert.Equal(t, tc.expected, b.Length())
		})
	}
}

func TestSampleBuffer_Read(t *testing.T) {
	data := []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	tt := []struct {
		desc         string
		read         int
		write        int
		count        int
		expectedRead int
		expectedData []float32
	}{
		{"empty", 0, 0, 10, 0, []float32{}},
		{"1", 0, 1, 1, 1, []float32{0}},
		{"1-3", 1, 5, 3, 4, []float32{1, 2, 3}},
		{"1-0", 1, 0, 10, 0, []float32{1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"8-2", 8, 7, 5, 3, []float32{8, 9, 0, 1, 2}},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			b := &sampleBuffer{read: tc.read, write: tc.write, samples: data}
			buf := make([]float32, tc.count)
			actualCount, err := b.Read(buf)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedRead, b.read, "read position")
			assert.Equal(t, len(tc.expectedData), actualCount, "actual count")
			assert.Equal(t, tc.expectedData, buf[0:len(tc.expectedData)], "data")
		})
	}
}

func TestSampleBuffer_Write(t *testing.T) {
	tt := []struct {
		desc          string
		read          int
		write         int
		data          []float32
		expectedRead  int
		expectedWrite int
		expectedData  []float32
	}{
		{"nothing", 0, 0, []float32{}, 0, 0, []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"1", 0, 5, []float32{10}, 0, 6, []float32{0, 1, 2, 3, 4, 10, 6, 7, 8, 9}},
		{"1 behind read", 6, 5, []float32{10}, 7, 6, []float32{0, 1, 2, 3, 4, 10, 6, 7, 8, 9}},
		{"1 around edge", 0, 9, []float32{10}, 1, 0, []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 10}},
		{"5 around edge", 5, 8, []float32{10, 11, 12, 13, 14}, 5, 3, []float32{12, 13, 14, 3, 4, 5, 6, 7, 10, 11}},
		{"> capacity", 0, 0, []float32{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, 3, 2, []float32{20, 21, 12, 13, 14, 15, 16, 17, 18, 19}},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			b := &sampleBuffer{read: tc.read, write: tc.write, samples: []float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}
			actualCount, err := b.Write(tc.data)
			require.NoError(t, err)
			assert.Equal(t, len(tc.data), actualCount, "actual count")
			assert.Equal(t, tc.expectedRead, b.read, "read position")
			assert.Equal(t, tc.expectedWrite, b.write, "write position")
			assert.Equal(t, tc.expectedData, b.samples, "data")
		})
	}
}
