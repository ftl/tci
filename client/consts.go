package client

import "fmt"

// VFO represents a VFO in TCI. In the TCI documentation the VFOs area also named "channel".
type VFO int

// All available VFOs in TCI.
const (
	VFOA = VFO(0)
	VFOB = VFO(1)
)

// Mode represents the modulation type of a TRX.
type Mode string

// All modes acvailable in TCI.
const (
	ModeAM   = Mode("am")
	ModeSAM  = Mode("sam")
	ModeDSB  = Mode("dsb")
	ModeLSB  = Mode("lsb")
	ModeUSB  = Mode("usb")
	ModeCW   = Mode("cw")
	ModeNFM  = Mode("nfm")
	ModeWFM  = Mode("wfm")
	ModeSPEC = Mode("spec")
	ModeDIGL = Mode("digl")
	ModeDIGU = Mode("digu")
	ModeDRM  = Mode("drm")
)

// SignalSource represents the source of the TX audio signal.
type SignalSource string

// All available sources for the TX audio signal.
const (
	SignalSourceDefault = SignalSource("")
	SignalSourceMIC     = SignalSource("mic")
	SignalSourceVAC     = SignalSource("vac")
)

// IQSampleRate represents the sample rate for IQ data in samples per second.
type IQSampleRate int

// All available sample rates for IQ data.
const (
	IQSampleRate48k  = IQSampleRate(48000)
	IQSampleRate96k  = IQSampleRate(96000)
	IQSampleRate192k = IQSampleRate(192000)
)

// AudioSampleRate represents the sample rate for audio data in samples per second.
type AudioSampleRate int

// All available sample rates for audio data.
const (
	AudioSampleRate8k  = AudioSampleRate(8000)
	AudioSampleRate12k = AudioSampleRate(12000)
	AudioSampleRate24k = AudioSampleRate(24000)
	AudioSampleRate48k = AudioSampleRate(48000)
)

// NewARGB returns a ARGB with the given alpha, red, green, and blue values.
func NewARGB(a, r, g, b byte) ARGB {
	return ARGB(uint32(a)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// ARGB represents a color with four (alpha, red, green, blue) 8-bit channels.
type ARGB uint32

func (c ARGB) String() string {
	return fmt.Sprintf("%d", uint32(c))
	// return fmt.Sprintf("#%x", uint32(c))
}

// A is the alpha channel of this color.
func (c ARGB) A() byte {
	return byte((c >> 24) & 0xff)
}

// R is the red channel of this color.
func (c ARGB) R() byte {
	return byte((c >> 16) & 0xff)
}

// G is the green channel of this color.
func (c ARGB) G() byte {
	return byte((c >> 8) & 0xff)
}

// B is the blue channel of this color.
func (c ARGB) B() byte {
	return byte(c & 0xff)
}

// CTCSSMode represents the operation mode for CTCSS subtones.
type CTCSSMode int

// All available CTCSS modes.
const (
	CTCSSModeBoth CTCSSMode = iota
	CTCSSModeRXOnly
	CTCSSModeTXOnly
)

// CTCSSTone represents a certain CTCSS subtone.
type CTCSSTone int
