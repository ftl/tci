package client

type DeviceInfo struct {
	Name            string
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

func (d *DeviceInfo) SetDeviceName(name string) {
	d.Name = name
}

func (d *DeviceInfo) SetProtocol(name string, version string) {
	d.ProtocolName = name
	d.ProtocolVersion = version
}

func (d *DeviceInfo) SetVFOLimits(min int, max int) {
	d.MinVFOFrequency = min
	d.MaxVFOFrequency = max
}

func (d *DeviceInfo) SetIFLimits(min int, max int) {
	d.MinIFFrequency = min
	d.MaxIFFrequency = max
}

func (d *DeviceInfo) SetTRXCount(count int) {
	d.TRXCount = count
}

func (d *DeviceInfo) SetChannelCount(count int) {
	d.ChannelCount = count
}

func (d *DeviceInfo) SetRXOnly(value bool) {
	d.RXOnly = value
}

func (d *DeviceInfo) SetModes(modes []Mode) {
	d.Modes = modes
}
