package client

import "log"

func newNotifier(listeners []interface{}, closed <-chan struct{}) *notifier {
	result := &notifier{
		listeners:      listeners,
		closed:         closed,
		textMessages:   make(chan Message, 1),
		binaryMessages: make(chan BinaryMessage, 1),
	}
	go result.notifyLoop()
	return result
}

type notifier struct {
	listeners      []interface{}
	closed         <-chan struct{}
	textMessages   chan Message
	binaryMessages chan BinaryMessage
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
	case "vfo_limits":
		err = n.emitVFOLimits(msg)
	case "if_limits":
		err = n.emitVFOLimits(msg)
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
	case "protocol":
		err = n.emitProtocol(msg)
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

type SquelchListener interface {
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
		if listener, ok := l.(SquelchListener); ok {
			listener.SetSquelchEnable(trx, enabled)
		}
	}
	return nil
}

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

type RXAudioListener interface {
	RXAudio(trx int, sampleRate AudioSampleRate, data []float32)
}

func (n *notifier) emitRXAudio(msg BinaryMessage) {
	for _, l := range n.listeners {
		if listener, ok := l.(RXAudioListener); ok {
			listener.RXAudio(msg.TRX, AudioSampleRate(msg.SampleRate), msg.Data)
		}
	}
}

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
