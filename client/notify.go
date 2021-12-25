package client

import "log"

const (
	tci_1_4 tciVersion = 1.4
	tci_1_5 tciVersion = 1.5
)

type tciVersion float64

// Beyond indicates if this TCI version is beyond the given version.
func (v tciVersion) Beyond(o tciVersion) bool {
	return v > o
}

// AtLeast indicates if this TCI version is at least the given version.
func (v tciVersion) AtLeast(o tciVersion) bool {
	return v >= o
}

func newNotifier(listeners []interface{}, closed <-chan struct{}) *notifier {
	result := &notifier{
		listeners:      listeners,
		closed:         closed,
		textMessages:   make(chan Message, 1),
		binaryMessages: make(chan BinaryMessage, 1),
		tciVersion:     1.4,
	}
	go result.notifyLoop()
	return result
}

type notifier struct {
	listeners      []interface{}
	closed         <-chan struct{}
	textMessages   chan Message
	binaryMessages chan BinaryMessage
	tciName        string
	tciVersion     tciVersion
}

func (n *notifier) notifyLoop() {
	for {
		select {
		case <-n.closed:
			return
		case msg := <-n.textMessages:
			n.handleIncomingMessage(msg)
		case msg := <-n.binaryMessages:
			n.handleIncomingBinaryMessage(msg)
		}
	}
}

// Notify registers the given listener. The listener is then notified about incoming messages.
func (n *notifier) Notify(listener interface{}) {
	n.listeners = append(n.listeners, listener)
}

func (n *notifier) textMessage(msg Message) {
	n.textMessages <- msg
}

func (n *notifier) handleIncomingMessage(msg Message) {
	n.emitMessage(msg)
	var err error
	switch msg.name {
	case "protocol":
		n.setTCIProtocol(msg)
		err = n.emitProtocol(msg)
	case "vfo_limits":
		err = n.emitVFOLimits(msg)
	case "if_limits":
		err = n.emitIFLimits(msg)
	case "trx_count":
		err = n.emitTRXCount(msg)
	case "channels_count":
		err = n.emitChannelCount(msg)
	case "device":
		err = n.emitDeviceName(msg)
	case "receive_only":
		err = n.emitRXOnly(msg)
	case "modulations_list":
		err = n.emitModes(msg)
	case "tx_enable":
		err = n.emitTXEnable(msg)
	case "ready":
		err = n.emitReady(msg)
	case "tx_footswitch":
		err = n.emitTXFootswitch(msg)
	case "start":
		err = n.emitStart(msg)
	case "stop":
		err = n.emitStop(msg)
	case "dds":
		err = n.emitDDS(msg)
	case "if":
		err = n.emitIF(msg)
	case "rit_enable":
		err = n.emitRITEnable(msg)
	case "modulation":
		err = n.emitMode(msg)
	case "rx_enable":
		err = n.emitRXEnable(msg)
	case "xit_enable":
		err = n.emitXITEnable(msg)
	case "split_enable":
		err = n.emitSplitEnable(msg)
	case "rit_offset":
		err = n.emitRITOffset(msg)
	case "xit_offset":
		err = n.emitXITOffset(msg)
	case "rx_channel_enable":
		err = n.emitRXChannelEnable(msg)
	case "rx_filter_band":
		err = n.emitRXFilterBand(msg)
	case "rx_smeter":
		err = n.emitRXSMeter(msg)
	case "cw_macros_speed":
		err = n.emitCWMacrosSpeed(msg)
	case "cw_macros_delay":
		err = n.emitCWMacrosDelay(msg)
	case "trx":
		err = n.emitTX(msg)
	case "tune":
		err = n.emitTune(msg)
	case "drive":
		err = n.emitDrive(msg)
	case "tune_drive":
		err = n.emitTuneDrive(msg)
	case "iq_start":
		err = n.emitStartIQ(msg)
	case "iq_stop":
		err = n.emitStopIQ(msg)
	case "iq_samplerate":
		err = n.emitIQSampleRate(msg)
	case "audio_start":
		err = n.emitStartAudio(msg)
	case "audio_stop":
		err = n.emitStopAudio(msg)
	case "audio_samplerate":
		err = n.emitAudioSampleRate(msg)
	case "tx_power":
		err = n.emitTXPower(msg)
	case "tx_swr":
		err = n.emitTXSWR(msg)
	case "volume":
		err = n.emitVolume(msg)
	case "sql_enable":
		err = n.emitSquelchEnable(msg)
	case "sql_level":
		err = n.emitSquelchLevel(msg)
	case "vfo":
		err = n.emitVFOFrequency(msg)
	case "app_focus":
		err = n.emitAppFocus(msg)
	case "mute":
		err = n.emitMute(msg)
	case "rx_mute":
		err = n.emitRXMute(msg)
	case "ctcss_enable":
		err = n.emitCTCSSEnable(msg)
	case "ctcss_mode":
		err = n.emitCTCSSMode(msg)
	case "ctcss_rx_tone":
		err = n.emitCTCSSRXTone(msg)
	case "ctcss_tx_tone":
		err = n.emitCTCSSTXTone(msg)
	case "ctcss_level":
		err = n.emitCTCSSLevel(msg)
	case "ecoder_switch_rx":
		err = n.emitECoderSwitchRX(msg)
	case "ecoder_switch_channel":
		err = n.emitECoderSwitchChannel(msg)
	case "rx_volume":
		err = n.emitRXVolume(msg)
	case "rx_balance":
		err = n.emitRXBalance(msg)
	default:
		log.Printf("unknown incoming message: %v", msg)
	}
	if err != nil {
		log.Printf("cannot emit message %v: %v", msg, err)
	}
}

func (n *notifier) setTCIProtocol(msg Message) {
	name, err := msg.ToString(0)
	if err != nil {
		log.Printf("Cannot parse protocol message: %w", err)
		return
	}
	version, err := msg.ToFloat(1)
	if err != nil {
		log.Printf("Cannot parse protocol version: %w", err)
		return
	}

	n.tciName = name
	n.tciVersion = tciVersion(version)
}

// MessageListener is notified when any text message is received from the TCI server.
type MessageListener interface {
	Message(msg Message)
}

func (n *notifier) emitMessage(msg Message) {
	for _, l := range n.listeners {
		if listener, ok := l.(MessageListener); ok {
			listener.Message(msg)
		}
	}
}

// A ProtocolListener is notified when a PROTOCOL message is received from the TCI server.
type ProtocolListener interface {
	SetProtocol(name string, version string)
}

func (n *notifier) emitProtocol(msg Message) error {
	name, err := msg.ToString(0)
	if err != nil {
		return err
	}
	version, err := msg.ToString(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ProtocolListener); ok {
			listener.SetProtocol(name, version)
		}
	}
	return nil
}

// A VFOLimitsListener is notified when a VFO_LIMITS message is received from the TCI server.
type VFOLimitsListener interface {
	SetVFOLimits(min, max int)
}

func (n *notifier) emitVFOLimits(msg Message) error {
	min, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	max, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(VFOLimitsListener); ok {
			listener.SetVFOLimits(min, max)
		}
	}
	return nil
}

// An IFLimitsListener is notified when an IF_LIMITS message is received from the TCI server.
type IFLimitsListener interface {
	SetIFLimits(min, max int)
}

func (n *notifier) emitIFLimits(msg Message) error {
	min, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	max, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(IFLimitsListener); ok {
			listener.SetIFLimits(min, max)
		}
	}
	return nil
}

// A TRXCountListener is notified when a TRX_COUNT message is received from the TCI server.
type TRXCountListener interface {
	SetTRXCount(count int)
}

func (n *notifier) emitTRXCount(msg Message) error {
	count, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TRXCountListener); ok {
			listener.SetTRXCount(count)
		}
	}
	return nil
}

// A ChannelCountListener is notified when a CHANNEL_COUNT message is received from the TCI server.
type ChannelCountListener interface {
	SetChannelCount(count int)
}

func (n *notifier) emitChannelCount(msg Message) error {
	count, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ChannelCountListener); ok {
			listener.SetChannelCount(count)
		}
	}
	return nil
}

// A DeviceNameListener is notified when a DEVICE message is received from the TCI server.
type DeviceNameListener interface {
	SetDeviceName(name string)
}

func (n *notifier) emitDeviceName(msg Message) error {
	name, err := msg.ToString(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(DeviceNameListener); ok {
			listener.SetDeviceName(name)
		}
	}
	return nil
}

// A RXOnlyListener is notified when a RECEIVE_ONLY message is received from the TCI server.
type RXOnlyListener interface {
	SetRXOnly(value bool)
}

func (n *notifier) emitRXOnly(msg Message) error {
	value, err := msg.ToBool(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXOnlyListener); ok {
			listener.SetRXOnly(value)
		}
	}
	return nil
}

// A ModesListener is notified when a MODULATIONS_LIST message is received from the TCI server.
type ModesListener interface {
	SetModes(modes []Mode)
}

func (n *notifier) emitModes(msg Message) error {
	modes := make([]Mode, len(msg.args))
	for i, arg := range msg.args {
		modes[i] = Mode(arg)
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ModesListener); ok {
			listener.SetModes(modes)
		}
	}
	return nil
}

// A TXEnableListener is notified when a TX_ENABLE message is received from the TCI server.
type TXEnableListener interface {
	SetTXEnable(trx int, enabled bool)
}

func (n *notifier) emitTXEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TXEnableListener); ok {
			listener.SetTXEnable(trx, enabled)
		}
	}
	return nil
}

// A ReadyListener is notified when a READY message is received from the TCI server.
type ReadyListener interface {
	Ready()
}

func (n *notifier) emitReady(Message) error {
	for _, l := range n.listeners {
		if listener, ok := l.(ReadyListener); ok {
			listener.Ready()
		}
	}
	return nil
}

// A TXFootswitchListener is notified when a TX_FOOTSWITCH message is received from the TCI server.
type TXFootswitchListener interface {
	SetTXFootswitch(trx int, pressed bool)
}

func (n *notifier) emitTXFootswitch(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	pressed, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TXFootswitchListener); ok {
			listener.SetTXFootswitch(trx, pressed)
		}
	}
	return nil
}

// A StartListener is notified when a START message is received from the TCI server.
type StartListener interface {
	Start()
}

func (n *notifier) emitStart(Message) error {
	for _, l := range n.listeners {
		if listener, ok := l.(StartListener); ok {
			listener.Start()
		}
	}
	return nil
}

// A StopListener is notified when a STOP message is received from the TCI server.
type StopListener interface {
	Stop()
}

func (n *notifier) emitStop(Message) error {
	for _, l := range n.listeners {
		if listener, ok := l.(StopListener); ok {
			listener.Stop()
		}
	}
	return nil
}

// A DDSListener is notified when a DDS message is received from the TCI server.
type DDSListener interface {
	SetDDS(trx int, frequency int)
}

func (n *notifier) emitDDS(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	frequency, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(DDSListener); ok {
			listener.SetDDS(trx, frequency)
		}
	}
	return nil
}

// An IFListener is notified when an IF_LIMITS message is received from the TCI server.
type IFListener interface {
	SetIF(trx int, vfo VFO, frequency int)
}

func (n *notifier) emitIF(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	frequency, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(IFListener); ok {
			listener.SetIF(trx, VFO(vfo), frequency)
		}
	}
	return nil
}

// A RITEnableListener is notified when a RIT_ENABLE message is received from the TCI server.
type RITEnableListener interface {
	SetRITEnable(trx int, enabled bool)
}

func (n *notifier) emitRITEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RITEnableListener); ok {
			listener.SetRITEnable(trx, enabled)
		}
	}
	return nil
}

// A ModeListener is notified when a MODULATION message is received from the TCI server.
type ModeListener interface {
	SetMode(trx int, mode Mode)
}

func (n *notifier) emitMode(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	mode, err := msg.ToString(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ModeListener); ok {
			listener.SetMode(trx, Mode(mode))
		}
	}
	return nil
}

// A RXEnableListener is notified when a RX_ENABLE message is received from the TCI server.
type RXEnableListener interface {
	SetRXEnable(trx int, enabled bool)
}

func (n *notifier) emitRXEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXEnableListener); ok {
			listener.SetRXEnable(trx, enabled)
		}
	}
	return nil
}

// A XITEnableListener is notified when a XIT_ENABLE message is received from the TCI server.
type XITEnableListener interface {
	SetXITEnable(trx int, enabled bool)
}

func (n *notifier) emitXITEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(XITEnableListener); ok {
			listener.SetXITEnable(trx, enabled)
		}
	}
	return nil
}

// A SplitEnableListener is notified when a SPLIT_ENABLE message is received from the TCI server.
type SplitEnableListener interface {
	SetSplitEnable(trx int, enabled bool)
}

func (n *notifier) emitSplitEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(SplitEnableListener); ok {
			listener.SetSplitEnable(trx, enabled)
		}
	}
	return nil
}

// A RITOffsetListener is notified when a RIT_OFFSET message is received from the TCI server.
type RITOffsetListener interface {
	SetRITOffset(trx int, offset int)
}

func (n *notifier) emitRITOffset(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	offset, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RITOffsetListener); ok {
			listener.SetRITOffset(trx, offset)
		}
	}
	return nil
}

// A XITOffsetListener is notified when a XIT_OFFSET message is received from the TCI server.
type XITOffsetListener interface {
	SetXITOffset(trx int, offset int)
}

func (n *notifier) emitXITOffset(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	offset, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(XITOffsetListener); ok {
			listener.SetXITOffset(trx, offset)
		}
	}
	return nil
}

// A RXChannelEnableListener is notified when a RX_CHANNEL_ENABLE message is received from the TCI server.
type RXChannelEnableListener interface {
	SetRXChannelEnable(trx int, vfo VFO, enabled bool)
}

func (n *notifier) emitRXChannelEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXChannelEnableListener); ok {
			listener.SetRXChannelEnable(trx, VFO(vfo), enabled)
		}
	}
	return nil
}

// A RXFilterBandListener is notified when a RX_FILTER_BAND message is received from the TCI server.
type RXFilterBandListener interface {
	SetRXFilterBand(trx int, min, max int)
}

func (n *notifier) emitRXFilterBand(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	min, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	max, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXFilterBandListener); ok {
			listener.SetRXFilterBand(trx, min, max)
		}
	}
	return nil
}

// A RXSMeterListener is notified when a RX_SMETER message is received from the TCI server.
type RXSMeterListener interface {
	SetRXSMeter(trx int, vfo VFO, level int)
}

func (n *notifier) emitRXSMeter(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	level, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXSMeterListener); ok {
			listener.SetRXSMeter(trx, VFO(vfo), level)
		}
	}
	return nil
}

// A CWMacrosSpeedListener is notified when a CW_MACROS_SPEED message is received from the TCI server.
type CWMacrosSpeedListener interface {
	SetCWMacrosSpeed(wpm int)
}

func (n *notifier) emitCWMacrosSpeed(msg Message) error {
	wpm, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CWMacrosSpeedListener); ok {
			listener.SetCWMacrosSpeed(wpm)
		}
	}
	return nil
}

// A CWMacrosDelayListener is notified when a CW_MACROS_DELAY message is received from the TCI server.
type CWMacrosDelayListener interface {
	SetCWMacrosDelay(delay int)
}

func (n *notifier) emitCWMacrosDelay(msg Message) error {
	delay, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CWMacrosDelayListener); ok {
			listener.SetCWMacrosDelay(delay)
		}
	}
	return nil
}

// A CWMacrosEmptyListener is notified then a CW_MACROS_EMPTY message is received from the TCI server. (since TCI 1.5)
type CWMacrosEmptyListener interface {
	CWMacrosEmpty()
}

func (n *notifier) emitCWMacrosEmpty(msg Message) error {
	for _, l := range n.listeners {
		if listener, ok := l.(CWMacrosEmptyListener); ok {
			listener.CWMacrosEmpty()
		}
	}
	return nil
}

// A TXListener is notified when a TRX message is received from the TCI server.
type TXListener interface {
	SetTX(trx int, enabled bool)
}

func (n *notifier) emitTX(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TXListener); ok {
			listener.SetTX(trx, enabled)
		}
	}
	return nil
}

// A TuneListener is notified when a TUNE message is received from the TCI server.
type TuneListener interface {
	SetTune(trx int, enabled bool)
}

func (n *notifier) emitTune(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TuneListener); ok {
			listener.SetTune(trx, enabled)
		}
	}
	return nil
}

// A DriveListener is notified when a DRIVE message is received from the TCI server.
type DriveListener interface {
	SetDrive(percent int)
}

func (n *notifier) emitDrive(msg Message) error {
	percent, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(DriveListener); ok {
			listener.SetDrive(percent)
		}
	}
	return nil
}

// A TuneDriveListener is notified when a TUNE_DRIVE message is received from the TCI server.
type TuneDriveListener interface {
	SetTuneDrive(percent int)
}

func (n *notifier) emitTuneDrive(msg Message) error {
	percent, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TuneDriveListener); ok {
			listener.SetTuneDrive(percent)
		}
	}
	return nil
}

// A StartIQListener is notified when a IQ_START message is received from the TCI server.
type StartIQListener interface {
	StartIQ(trx int)
}

func (n *notifier) emitStartIQ(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(StartIQListener); ok {
			listener.StartIQ(trx)
		}
	}
	return nil
}

// A StopIQListener is notified when a IQ_STOP message is received from the TCI server.
type StopIQListener interface {
	StopIQ(trx int)
}

func (n *notifier) emitStopIQ(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(StopIQListener); ok {
			listener.StopIQ(trx)
		}
	}
	return nil
}

// A IQSampleRateListener is notified when a IQ_SAMPLERATE message is received from the TCI server.
type IQSampleRateListener interface {
	SetIQSampleRate(sampleRate IQSampleRate)
}

func (n *notifier) emitIQSampleRate(msg Message) error {
	sampleRate, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(IQSampleRateListener); ok {
			listener.SetIQSampleRate(IQSampleRate(sampleRate))
		}
	}
	return nil
}

// A StartAudioListener is notified when a AUDIO_START message is received from the TCI server.
type StartAudioListener interface {
	StartAudio(trx int)
}

func (n *notifier) emitStartAudio(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(StartAudioListener); ok {
			listener.StartAudio(trx)
		}
	}
	return nil
}

// A StopAudioListener is notified when a AUDIO_STOP message is received from the TCI server.
type StopAudioListener interface {
	StopAudio(trx int)
}

func (n *notifier) emitStopAudio(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(StopAudioListener); ok {
			listener.StopAudio(trx)
		}
	}
	return nil
}

// A AudioSampleRateListener is notified when a AUDIO_SAMPLERATE message is received from the TCI server.
type AudioSampleRateListener interface {
	SetAudioSampleRate(sampleRate AudioSampleRate)
}

func (n *notifier) emitAudioSampleRate(msg Message) error {
	sampleRate, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(AudioSampleRateListener); ok {
			listener.SetAudioSampleRate(AudioSampleRate(sampleRate))
		}
	}
	return nil
}

// A TXPowerListener is notified when a TX_POWER message is received from the TCI server.
type TXPowerListener interface {
	SetTXPower(watts float64)
}

func (n *notifier) emitTXPower(msg Message) error {
	watts, err := msg.ToFloat(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TXPowerListener); ok {
			listener.SetTXPower(watts)
		}
	}
	return nil
}

// A TXSWRListener is notified when a TX_SWR message is received from the TCI server.
type TXSWRListener interface {
	SetTXSWR(ratio float64)
}

func (n *notifier) emitTXSWR(msg Message) error {
	ratio, err := msg.ToFloat(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(TXSWRListener); ok {
			listener.SetTXSWR(ratio)
		}
	}
	return nil
}

// A VolumeListener is notified when a VOLUME message is received from the TCI server.
type VolumeListener interface {
	SetVolume(dB int)
}

func (n *notifier) emitVolume(msg Message) error {
	dB, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(VolumeListener); ok {
			listener.SetVolume(dB)
		}
	}
	return nil
}

// A SquelchEnableListener is notified when a SQL_ENABLE message is received from the TCI server.
type SquelchEnableListener interface {
	SetSquelchEnable(trx int, enabled bool)
}

func (n *notifier) emitSquelchEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(SquelchEnableListener); ok {
			listener.SetSquelchEnable(trx, enabled)
		}
	}
	return nil
}

// A SquelchLevelListener is notified when a SQL_LEVEL message is received from the TCI server.
type SquelchLevelListener interface {
	SetSquelchLevel(dB int)
}

func (n *notifier) emitSquelchLevel(msg Message) error {
	dB, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(SquelchLevelListener); ok {
			listener.SetSquelchLevel(dB)
		}
	}
	return nil
}

// A VFOFrequencyListener is notified when a VFO message is received from the TCI server.
type VFOFrequencyListener interface {
	SetVFOFrequency(trx int, vfo VFO, frequency int)
}

func (n *notifier) emitVFOFrequency(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	frequency, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(VFOFrequencyListener); ok {
			listener.SetVFOFrequency(trx, VFO(vfo), frequency)
		}
	}
	return nil
}

// A AppFocusListener is notified when a APP_FOCUS message is received from the TCI server.
type AppFocusListener interface {
	SetAppFocus(focussed bool)
}

func (n *notifier) emitAppFocus(msg Message) error {
	focussed, err := msg.ToBool(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(AppFocusListener); ok {
			listener.SetAppFocus(focussed)
		}
	}
	return nil
}

// A MuteListener is notified when a MUTE message is received from the TCI server.
type MuteListener interface {
	SetMute(muted bool)
}

func (n *notifier) emitMute(msg Message) error {
	muted, err := msg.ToBool(0)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(MuteListener); ok {
			listener.SetMute(muted)
		}
	}
	return nil
}

// A RXMuteListener is notified when a RX_MUTE message is received from the TCI server.
type RXMuteListener interface {
	SetRXMute(trx int, muted bool)
}

func (n *notifier) emitRXMute(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	muted, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXMuteListener); ok {
			listener.SetRXMute(trx, muted)
		}
	}
	return nil
}

// A CTCSSEnableListener is notified when a CTCSS_ENABLE message is received from the TCI server.
type CTCSSEnableListener interface {
	SetCTCSSEnable(trx int, enabled bool)
}

func (n *notifier) emitCTCSSEnable(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	enabled, err := msg.ToBool(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CTCSSEnableListener); ok {
			listener.SetCTCSSEnable(trx, enabled)
		}
	}
	return nil
}

// A CTCSSModeListener is notified when a CTCSS_MODE message is received from the TCI server.
type CTCSSModeListener interface {
	SetCTCSSMode(trx int, mode CTCSSMode)
}

func (n *notifier) emitCTCSSMode(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	mode, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CTCSSModeListener); ok {
			listener.SetCTCSSMode(trx, CTCSSMode(mode))
		}
	}
	return nil
}

// A CTCSSRXToneListener is notified when a CTCSS_RX_TONE message is received from the TCI server.
type CTCSSRXToneListener interface {
	SetCTCSSRXTone(trx int, tone CTCSSTone)
}

func (n *notifier) emitCTCSSRXTone(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	tone, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CTCSSRXToneListener); ok {
			listener.SetCTCSSRXTone(trx, CTCSSTone(tone))
		}
	}
	return nil
}

// A CTCSSTXToneListener is notified when a CTCSS_TX_TONE message is received from the TCI server.
type CTCSSTXToneListener interface {
	SetCTCSSTXTone(trx int, tone CTCSSTone)
}

func (n *notifier) emitCTCSSTXTone(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	tone, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CTCSSTXToneListener); ok {
			listener.SetCTCSSTXTone(trx, CTCSSTone(tone))
		}
	}
	return nil
}

// A CTCSSLevelListener is notified when a CTCSS_LEVEL message is received from the TCI server.
type CTCSSLevelListener interface {
	SetCTCSSLevel(trx int, percent int)
}

func (n *notifier) emitCTCSSLevel(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	percent, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(CTCSSLevelListener); ok {
			listener.SetCTCSSLevel(trx, percent)
		}
	}
	return nil
}

// A ECoderSwitchRXListener is notified when a EXODER_SWITCH_RX message is received from the TCI server.
type ECoderSwitchRXListener interface {
	SetECoderSwitchRX(ecoder int, trx int)
}

func (n *notifier) emitECoderSwitchRX(msg Message) error {
	ecoder, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	trx, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ECoderSwitchRXListener); ok {
			listener.SetECoderSwitchRX(ecoder, trx)
		}
	}
	return nil
}

// A ECoderSwitchChannelListener is notified when a ECODER_SWITCH_CHANNEL message is received from the TCI server.
type ECoderSwitchChannelListener interface {
	SetECoderSwitchChannel(ecoder int, vfo VFO)
}

func (n *notifier) emitECoderSwitchChannel(msg Message) error {
	ecoder, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(ECoderSwitchChannelListener); ok {
			listener.SetECoderSwitchChannel(ecoder, VFO(vfo))
		}
	}
	return nil
}

// A RXVolumeListener is notified when a RX_VOLUME message is received from the TCI server.
type RXVolumeListener interface {
	SetRXVolume(trx int, vfo VFO, dB int)
}

func (n *notifier) emitRXVolume(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	dB, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXVolumeListener); ok {
			listener.SetRXVolume(trx, VFO(vfo), dB)
		}
	}
	return nil
}

// A RXBalanceListener is notified when a RX_BALANCE message is received from the TCI server.
type RXBalanceListener interface {
	SetRXBalance(trx int, vfo VFO, dB int)
}

func (n *notifier) emitRXBalance(msg Message) error {
	trx, err := msg.ToInt(0)
	if err != nil {
		return err
	}
	vfo, err := msg.ToInt(1)
	if err != nil {
		return err
	}
	dB, err := msg.ToInt(2)
	if err != nil {
		return err
	}
	for _, l := range n.listeners {
		if listener, ok := l.(RXBalanceListener); ok {
			listener.SetRXBalance(trx, VFO(vfo), dB)
		}
	}
	return nil
}

func (n *notifier) binaryMessage(msg BinaryMessage) {
	n.binaryMessages <- msg
}

func (n *notifier) handleIncomingBinaryMessage(msg BinaryMessage) {
	n.emitBinaryMessage(msg)
	switch msg.Type {
	case IQStreamMessage:
		n.emitIQData(msg)
	case RXAudioStreamMessage:
		n.emitRXAudio(msg)
	case TXChronoMessage:
		n.emitTXChrono(msg)
	default:
		log.Printf("unknown binary message type: %v", msg.Type)
	}
}

// A BinaryMessageListener is notified when any binary message (IQ data, audio data, or tx chrono) is received from the TCI server.
type BinaryMessageListener interface {
	BinaryMessage(msg BinaryMessage)
}

func (n *notifier) emitBinaryMessage(msg BinaryMessage) {
	for _, l := range n.listeners {
		if listener, ok := l.(BinaryMessageListener); ok {
			listener.BinaryMessage(msg)
		}
	}
}

// A IQDataListener is notified when IQ data is received from the TCI server.
type IQDataListener interface {
	IQData(trx int, sampleRate IQSampleRate, data []float32)
}

func (n *notifier) emitIQData(msg BinaryMessage) {
	for _, l := range n.listeners {
		if listener, ok := l.(IQDataListener); ok {
			listener.IQData(msg.TRX, IQSampleRate(msg.SampleRate), msg.Data)
		}
	}
}

// A RXAudioListener is notified when RX audio data is received from the TCI server.
type RXAudioListener interface {
	RXAudio(trx int, sampleRate AudioSampleRate, samples []float32)
}

func (n *notifier) emitRXAudio(msg BinaryMessage) {
	for _, l := range n.listeners {
		if listener, ok := l.(RXAudioListener); ok {
			listener.RXAudio(msg.TRX, AudioSampleRate(msg.SampleRate), msg.Data)
		}
	}
}

// A TXChronoListener is notified when a TX chrono message is received from the TCI server.
type TXChronoListener interface {
	TXChrono(trx int, sampleRate AudioSampleRate, requestedSampleCount uint32)
}

func (n *notifier) emitTXChrono(msg BinaryMessage) {
	for _, l := range n.listeners {
		if listener, ok := l.(TXChronoListener); ok {
			listener.TXChrono(msg.TRX, AudioSampleRate(msg.SampleRate), msg.DataLength)
		}
	}
}
