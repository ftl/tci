package client

import (
	"fmt"
	"log"
	"sync"
)

const (
	rxBufferSize = 2 * 8192
	txBufferSize = 2 * 4096
	iqBufferSize = 2 * 4096
)

func newRXAudioStream(trx int, sampleRate AudioSampleRate, bufferSize int, closer func()) *RXAudioStream {
	return &RXAudioStream{
		closer:     closer,
		trx:        trx,
		sampleRate: sampleRate,
		closed:     make(chan struct{}),
		rxBuffer:   newSampleBuffer(bufferSize),
		rxWait:     make(chan bool, 1),
	}
}

type RXAudioStream struct {
	closer     func()
	trx        int
	sampleRate AudioSampleRate
	closed     chan struct{}
	rxBuffer   *sampleBuffer
	rxWait     chan bool
}

func (s *RXAudioStream) RXAudio(trx int, sampleRate AudioSampleRate, samples []float32) {
	if trx != s.trx {
		return
	}
	select {
	case <-s.closed:
		return
	default:
	}

	s.rxBuffer.Write(samples)
	select {
	case s.rxWait <- true:
	default:
	}
}

func (s *RXAudioStream) Read(to []float32) (int, error) {
	if !s.rxBuffer.HasNext() {
		<-s.rxWait
	} else {
		select {
		case <-s.rxWait:
		default:
		}
	}
	return s.rxBuffer.Read(to)
}

func (s *RXAudioStream) SampleRate() AudioSampleRate {
	return s.sampleRate
}

func (s *RXAudioStream) Close() error {
	select {
	case <-s.closed:
		return nil
	default:
		s.closer()
		close(s.closed)
		return nil
	}
}

func newStreamer(notifier *notifier, controller streamController) *streamer {
	result := &streamer{
		controller:    controller,
		rxStreams:     make(map[int]map[int]*RXAudioStream),
		rxStreamMutex: new(sync.RWMutex),
	}
	notifier.Notify(result)
	return result
}

type streamer struct {
	controller streamController

	nextRXStream  int
	rxStreams     map[int]map[int]*RXAudioStream
	rxStreamMutex *sync.RWMutex
}

type streamController interface {
	StartAudio(trx int) error
	StopAudio(trx int) error
	AudioSampleRate() (AudioSampleRate, error)
}

func (s *streamer) Close() {
	s.rxStreamMutex.Lock()
	defer s.rxStreamMutex.Unlock()

	for trx, trxStreams := range s.rxStreams {
		for _, stream := range trxStreams {
			stream.Close()
		}
		delete(s.rxStreams, trx)
	}
}

func (s *streamer) NewRXAudioStream(trx int) (*RXAudioStream, error) {
	s.rxStreamMutex.Lock()
	defer s.rxStreamMutex.Unlock()

	sampleRate, err := s.controller.AudioSampleRate()
	if err != nil {
		return nil, fmt.Errorf("cannot get audio sample rate: %w", err)
	}
	if sampleRate == 0 {
		return nil, fmt.Errorf("invalid audio sample rate: %d", sampleRate)
	}

	id := s.nextRXStream
	s.nextRXStream++
	stream := newRXAudioStream(trx, sampleRate, rxBufferSize, func() { s.closeRXAudioStream(trx, id) })
	trxStreams, ok := s.rxStreams[trx]
	if !ok {
		trxStreams = make(map[int]*RXAudioStream)
	}
	trxStreams[id] = stream
	s.rxStreams[trx] = trxStreams

	return stream, nil
}

func (s *streamer) closeRXAudioStream(trx int, id int) {
	s.rxStreamMutex.Lock()
	defer s.rxStreamMutex.Unlock()

	trxStreams, ok := s.rxStreams[trx]
	if !ok {
		return
	}

	delete(trxStreams, id)
}

func (s *streamer) RXAudio(trx int, sampleRate AudioSampleRate, samples []float32) {
	s.rxStreamMutex.RLock()
	defer s.rxStreamMutex.RUnlock()
	trxStreams, ok := s.rxStreams[trx]
	if !ok {
		return
	}
	for _, stream := range trxStreams {
		stream.RXAudio(trx, sampleRate, samples)
	}
}

/*
	SampleBuffer
*/

func newSampleBuffer(capacity int) *sampleBuffer {
	if capacity < 1 {
		panic("sampleBuffer must have a capacity > 0")
	}
	return &sampleBuffer{samples: make([]float32, capacity+1)}
}

type sampleBuffer struct {
	read    int
	write   int
	samples []float32
}

func (b sampleBuffer) String() string {
	return fmt.Sprintf("buffer: read %04d write %04d", b.read, b.write)
}

func (b *sampleBuffer) Length() int {
	switch {
	case b.write == b.read:
		return 0
	case b.write < b.read:
		return len(b.samples) - (b.read - b.write)
	default:
		return b.write - b.read
	}
}

func (b *sampleBuffer) Read(to []float32) (int, error) {
	capacity := len(b.samples)
	count := len(to)
	if count > b.Length() {
		count = b.Length()
	}

	if b.read+count < capacity {
		copy(to, b.samples[b.read:b.read+count])
		b.read += count
		return count, nil
	}

	pivot := capacity - b.read
	remainder := count - pivot
	copy(to, b.samples[b.read:capacity])
	copy(to[pivot:], b.samples[0:remainder])
	b.read = remainder

	return count, nil
}

func (b *sampleBuffer) HasNext() bool {
	return b.read != b.write
}

func (b *sampleBuffer) PeekNext() float32 {
	if b.read == b.write {
		return 0
	}
	value := b.samples[b.read]
	return value
}

func (b *sampleBuffer) PopNext() float32 {
	if b.read == b.write {
		return 0
	}
	value := b.samples[b.read]
	b.read = (b.read + 1) % len(b.samples)
	return value
}

func (b *sampleBuffer) Write(from []float32) (int, error) {
	capacity := len(b.samples)
	count := len(from)
	newWrite := (b.write + count) % capacity

	if count > capacity {
		pivot := count - newWrite
		copy(b.samples[0:newWrite], from[pivot:count])
		copy(b.samples[newWrite:], from[pivot-(capacity-newWrite):pivot])

		log.Printf("buffer overflow, dropping %d samples (0)", newWrite-b.read)
		b.read = (newWrite + 1) % capacity
		b.write = newWrite
		return count, nil
	}

	if b.write+count >= capacity {
		pivot := capacity - b.write
		copy(b.samples[b.write:], from[0:pivot])
		copy(b.samples, from[pivot:])

		if newWrite >= b.read {
			log.Printf("buffer overflow, dropping %d samples (1)", newWrite-b.read)
			b.read = (newWrite + 1) % capacity
		}
		b.write = newWrite
		return count, nil
	}

	copy(b.samples[b.write:], from)

	if b.write < b.read && newWrite >= b.read {
		log.Printf("buffer overflow, dropping %d samples (2)", newWrite-b.read)
		b.read = (newWrite + 1) % capacity
	}
	b.write = newWrite
	return count, nil
}
