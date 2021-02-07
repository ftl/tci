package client

// DeviceInfo contains the basic information about the SunSDR device which are transmitted when the TCI connection is established.
// This information cannot be requested through TCI, so it is automatically collected by the client and provided through this type.
type DeviceInfo struct {
	DeviceName      string
	ProtocolName    string
	ProtocolVersion string
	MinVFOFrequency int
	MaxVFOFrequency int
	MinIFFrequency  int
	MaxIFFrequency  int
	TRXCount        int
	ChannelCount    int
	RXOnly          bool
	Modes           []Mode
}

// SetDeviceName handles the DEVICE message.
func (d *DeviceInfo) SetDeviceName(name string) {
	d.DeviceName = name
}

// SetProtocol handles the PROTOCOL message.
func (d *DeviceInfo) SetProtocol(name string, version string) {
	d.ProtocolName = name
	d.ProtocolVersion = version
}

// SetVFOLimits handles the VFO_LIMITS message.
func (d *DeviceInfo) SetVFOLimits(min int, max int) {
	d.MinVFOFrequency = min
	d.MaxVFOFrequency = max
}

// SetIFLimits handles the IF_LIMITS message.
func (d *DeviceInfo) SetIFLimits(min int, max int) {
	d.MinIFFrequency = min
	d.MaxIFFrequency = max
}

// SetTRXCount handles the TRX_COUNT message.
func (d *DeviceInfo) SetTRXCount(count int) {
	d.TRXCount = count
}

// SetChannelCount handles the CHANNEL_COUNT message.
func (d *DeviceInfo) SetChannelCount(count int) {
	d.ChannelCount = count
}

// SetRXOnly handles the RECEIVE_ONLY message.
func (d *DeviceInfo) SetRXOnly(value bool) {
	d.RXOnly = value
}

// SetModes handles the MODULATIONS_LIST message.
func (d *DeviceInfo) SetModes(modes []Mode) {
	d.Modes = modes
}
