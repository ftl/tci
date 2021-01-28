package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// DefaultPort of TCI
const DefaultPort = 40001

// ReadTimeout is the duration to wait for a reply to a reading command.
var ReadTimeout = time.Duration(200 * time.Millisecond)

// ErrReadTimeout indicates a timeout while waiting for a reply to reading command.
var ErrReadTimeout = errors.New("read timeout")

// ErrNotConnected indicates that there is currently no TCI connection available.
var ErrNotConnected = errors.New("not connected")

// Client represents a TCI client.
type Client struct {
	notifier
	host           *net.TCPAddr
	closed         chan struct{}
	disconnectChan chan struct{}
	writeChan      chan command
}

type command struct {
	Message
	reply chan reply
}

type reply struct {
	Message
	err error
}

type clientConn interface {
	RemoteAddr() net.Addr
	Close() error
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
}

// Open a connection to the given host
func Open(host *net.TCPAddr) (*Client, error) {
	client := Client{
		host:   host,
		closed: make(chan struct{}),
	}
	err := client.connect()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func KeepOpen(host *net.TCPAddr, retryInterval time.Duration) *Client {
	client := &Client{
		host:   host,
		closed: make(chan struct{}),
	}
	go func() {
		disconnected := make(chan bool, 1)
		for {
			err := client.connect()
			if err == nil {
				client.WhenDisconnected(func() {
					disconnected <- true
				})
				select {
				case <-disconnected:
					log.Printf("connection lost to %s, waiting for retry", host.IP.String())
				case <-client.closed:
					log.Printf("connection closed")
					return
				}
			} else {
				log.Printf("cannot connect to %s, waiting for retry: %v", host.IP.String(), err)
			}

			select {
			case <-time.After(retryInterval):
				log.Printf("retrying to connect to %s", host.IP.String())
			case <-client.closed:
				log.Print("connection closed")
				return
			}
		}
	}()
	return client
}

func (c *Client) connect() error {
	if c.Connected() {
		return nil
	}

	host := c.host.IP.String()
	port := c.host.Port
	if port == 0 {
		port = DefaultPort
	}

	u, err := url.Parse(fmt.Sprintf("ws://%s:%d", host, port))
	if err != nil {
		return err
	}
	u.Port()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("cannot open websocket connection: %w", err)
	}
	c.disconnectChan = make(chan struct{})
	c.writeChan = make(chan command, 1)
	remoteAddr := conn.RemoteAddr()
	log.Printf("connected to %s", remoteAddr.String())

	incoming := make(chan Message, 1)

	go c.readLoop(conn, incoming)
	go c.writeLoop(conn, incoming)

	return nil
}

func (c *Client) readLoop(conn clientConn, incoming chan<- Message) {
	defer conn.Close()
	for {
		select {
		case <-c.disconnectChan:
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("cannot read next message: %v", err)
				close(c.disconnectChan)
				return
			}
			switch msgType {
			case websocket.TextMessage:
				message, err := ParseMessage(string(msg))
				if err != nil {
					log.Printf("cannot parse incoming message: %v", err)
					continue
				}
				c.handleIncomingMessage(message)
				incoming <- message
			default:
				log.Printf("unknown message type: %d %v", msgType, msg)
			}
		}
	}
}

func (c *Client) writeLoop(conn clientConn, incoming <-chan Message) {
	defer conn.Close()

	var currentCommand *command
	var currentDeadline time.Time

	for {
		now := time.Now()
		select {
		case <-c.disconnectChan:
			return
		case msg := <-incoming:
			if currentCommand == nil {
				continue
			}
			if msg.IsReplyTo(currentCommand.Message) {
				currentCommand.reply <- reply{Message: msg}
				currentCommand = nil
			}
		default:
			if currentCommand == nil {
				select {
				case cmd := <-c.writeChan:
					if cmd.reply != nil {
						currentCommand = &cmd
						currentDeadline = now.Add(ReadTimeout)
					}
					err := conn.WriteMessage(websocket.TextMessage, []byte(cmd.String()))
					if err != nil {
						log.Printf("error writing command %q: %v", cmd, err)
						continue
					}
				default:
					continue
				}
			} else if now.After(currentDeadline) {
				currentCommand.reply <- reply{err: ErrReadTimeout}
				currentCommand = nil
			}
		}
	}
}

func (c *Client) Connected() bool {
	if c.disconnectChan == nil {
		return false
	}
	select {
	case <-c.disconnectChan:
		return false
	default:
		return true
	}
}

func (c *Client) Disconnect() {
	// When the connection was disconnected from the outside, we keep it closed.
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}

	if c.disconnectChan == nil {
		return
	}
	select {
	case <-c.disconnectChan:
		return
	default:
		close(c.disconnectChan)
	}
}

func (c *Client) WhenDisconnected(f func()) {
	if c.disconnectChan == nil {
		f()
		return
	}
	go func() {
		<-c.disconnectChan
		f()
	}()
}

func (c *Client) command(cmd string, args ...interface{}) (Message, error) {
	if !c.Connected() {
		return Message{}, ErrNotConnected
	}
	replyChan := make(chan reply, 1)
	c.writeChan <- command{
		Message: NewMessage(cmd, args...),
		reply:   replyChan,
	}
	reply := <-replyChan

	return reply.Message, reply.err
}

// Start ExpertSDR2.
func (c *Client) Start() error {
	_, err := c.command("start")
	return err
}

// Stop ExpertSDR2.
func (c *Client) Stop() error {
	_, err := c.command("stop")
	return err
}

// SetDDS sets the center frequency of the given TRX's panorama.
func (c *Client) SetDDS(trx int, frequency int) error {
	_, err := c.command("dds", trx, frequency)
	return err
}

// DDS reads the center frequency of the given TRX's panorama.
func (c *Client) DDS(trx int) (int, error) {
	reply, err := c.command("dds", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetIF sets the tuning frequency of the given TRX's vfo.
func (c *Client) SetIF(trx int, vfo VFO, frequency int) error {
	_, err := c.command("if", trx, vfo, frequency)
	return err
}

// IF reads the tuning frequency of the given TRX's vfo.
func (c *Client) IF(trx int, vfo VFO) (int, error) {
	reply, err := c.command("if", trx, vfo)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(2)
}

// SetRITEnable enables the RIT of the given TRX.
func (c *Client) SetRITEnable(trx int, enabled bool) error {
	_, err := c.command("rit_enable", trx, enabled)
	return err
}

// RITEnable reads the RIT enable state of the given TRX.
func (c *Client) RITEnable(trx int) (bool, error) {
	reply, err := c.command("rit_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetMode sets the mode of the given TRX.
func (c *Client) SetMode(trx int, mode Mode) error {
	_, err := c.command("modulation", trx, mode)
	return err
}

// Mode reads the mode of the given TRX.
func (c *Client) Mode(trx int) (Mode, error) {
	reply, err := c.command("modulation", trx)
	if err != nil {
		return "", err
	}
	mode, err := reply.ToString(1)
	if err != nil {
		return "", err
	}
	return Mode(mode), nil
}

// SetRXEnable enables the RX of the given TRX.
func (c *Client) SetRXEnable(trx int, enabled bool) error {
	_, err := c.command("rx_enable", trx, enabled)
	return err
}

// RXEnable reads the RX enable state of the given TRX.
func (c *Client) RXEnable(trx int) (bool, error) {
	reply, err := c.command("rx_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetXITEnable enables the XIT of the given TRX.
func (c *Client) SetXITEnable(trx int, enabled bool) error {
	_, err := c.command("xit_enable", trx, enabled)
	return err
}

// XITEnable reads the XIT enable state of the given TRX.
func (c *Client) XITEnable(trx int) (bool, error) {
	reply, err := c.command("xit_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetSplitEnable enables the split mode of the given TRX.
func (c *Client) SetSplitEnable(trx int, enabled bool) error {
	_, err := c.command("split_enable", trx, enabled)
	return err
}

// SplitEnable reads the split mode enable state of the given TRX.
func (c *Client) SplitEnable(trx int) (bool, error) {
	reply, err := c.command("split_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetRITOffset sets the RIT offset in Hz for the given TRX.
func (c *Client) SetRITOffset(trx int, offset int) error {
	_, err := c.command("rit_offset", trx, offset)
	return err
}

// RITOffset reads the RIT offset in Hz for the given TRX.
func (c *Client) RITOffset(trx int) (int, error) {
	reply, err := c.command("rit_offset", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetXITOffset sets the XIT offset in Hz for the given TRX.
func (c *Client) SetXITOffset(trx int, offset int) error {
	_, err := c.command("xit_offset", trx, offset)
	return err
}

// XITOffset reads the XIT offset in Hz for the given TRX.
func (c *Client) XITOffset(trx int) (int, error) {
	reply, err := c.command("xit_offset", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetRXChannelEnable enables the given TRX's additional RX channel with the given index.
func (c *Client) SetRXChannelEnable(trx int, vfo VFO, enabled bool) error {
	_, err := c.command("rx_channel_enable", trx, vfo, enabled)
	return err
}

// RXChannelEnable reads the enable state of the given TRX's additional RX channel with the given index.
func (c *Client) RXChannelEnable(trx int, vfo VFO) (bool, error) {
	reply, err := c.command("rx_channel_enable", trx, vfo)
	if err != nil {
		return false, err
	}
	return reply.ToBool(2)
}

// SetRXFilterBand sets the IF filter boundaries of the given TRX using the given limit frequencies in Hz.
func (c *Client) SetRXFilterBand(trx int, min, max int) error {
	_, err := c.command("rx_filter_band", trx, min, max)
	return err
}

// RXFilterBand reads the IF filter boundaries of the given TRX.
func (c *Client) RXFilterBand(trx int) (int, int, error) {
	reply, err := c.command("rx_filter_band")
	if err != nil {
		return 0, 0, err
	}
	min, err := reply.ToInt(1)
	if err != nil {
		return 0, 0, err
	}
	max, err := reply.ToInt(2)
	if err != nil {
		return 0, 0, err
	}
	return min, max, nil
}

// SetRXSMeter sets the signal level of the given TRX's RX channel with the given index.
func (c *Client) SetRXSMeter(trx int, vfo VFO, level int) error {
	_, err := c.command("rx_smeter", trx, vfo, level)
	return err
}

// RXSMeter reads the signal level of the given TRX's RX channel with the given index.
func (c *Client) RXSMeter(trx int, vfo VFO) (int, error) {
	reply, err := c.command("rx_smeter", trx, vfo)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(2)
}

// SetCWMacrosSpeed sets the speed in WPM for CW macros.
func (c *Client) SetCWMacrosSpeed(wpm int) error {
	_, err := c.command("cw_macros_speed", wpm)
	return err
}

// CWMacrosSpeed reads the speed in WPM for CW macros.
func (c *Client) CWMacrosSpeed() (int, error) {
	reply, err := c.command("cw_macros_speed")
	if err != nil {
		return 0, err
	}
	return reply.ToInt(0)
}

// CWMacrosSpeedInc increases the speed for CW macros by the given delta in WPM.
func (c *Client) CWMacrosSpeedInc(delta int) error {
	_, err := c.command("cw_macros_speed_up", delta)
	return err
}

// CWMacrosSpeedDec decreases the speed for CW macros by the given delta in WPM.
func (c *Client) CWMacrosSpeedDec(delta int) error {
	_, err := c.command("cw_macros_speed_down", delta)
	return err
}

// SetCWMacrosDelay sets the delay between keying the TRX and transmitting a CW macros in milliseconds.
func (c *Client) SetCWMacrosDelay(delay int) error {
	_, err := c.command("cw_macros_delay", delay)
	return err
}

// CWMacrosDelay reads the delay for transmitting CW macros in milliseconds.
func (c *Client) CWMacrosDelay() (int, error) {
	reply, err := c.command("cw_macros_delay")
	if err != nil {
		return 0, err
	}
	return reply.ToInt(0)
}

// SetTX enables the TX of the given TRX using the given signal source. Use "" (SignalSourceDefault) if you want to use the default source for the current mode.
func (c *Client) SetTX(trx int, enabled bool, source SignalSource) error {
	var err error
	if source == SignalSourceDefault {
		_, err = c.command("trx", trx, enabled)
	} else {
		_, err = c.command("trx", trx, enabled, source)
	}
	return err
}

// TX reads the current state of the given TRX's transmitter.
func (c *Client) TX(trx int) (bool, error) {
	reply, err := c.command("trx", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetTune enables the given TRX's transmitter in tuning.
func (c *Client) SetTune(trx int, enabled bool) error {
	_, err := c.command("tune", trx, enabled)
	return err
}

// Tune reads the current state of the given TRX's tuning transmitter.
func (c *Client) Tune(trx int) (bool, error) {
	reply, err := c.command("tune", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetDrive sets the given TRX's output power in percent.
func (c *Client) SetDrive(trx int, percent int) error {
	_, err := c.command("drive", trx, percent)
	return err
}

// Drive reads the given TRX's output power in percent.
func (c *Client) Drive(trx int) (int, error) {
	reply, err := c.command("drive", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetTuneDrive sets the given TRX's output power in percent.
func (c *Client) SetTuneDrive(trx int, percent int) error {
	_, err := c.command("tune_drive", trx, percent)
	return err
}

// TuneDrive reads the given TRX's output power in percent.
func (c *Client) TuneDrive(trx int) (int, error) {
	reply, err := c.command("tune_drive", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// StartIQ starts the transmission of IQ data for the given TRX.
func (c *Client) StartIQ(trx int) error {
	_, err := c.command("iq_start", trx)
	return err
}

// StopIQ stops the transmission of IQ data for the given TRX.
func (c *Client) StopIQ(trx int) error {
	_, err := c.command("iq_stop", trx)
	return err
}

// SetIQSampleRate sets sample rate for IQ data.
func (c *Client) SetIQSampleRate(sampleRate IQSampleRate) error {
	_, err := c.command("iq_samplerate", sampleRate)
	return err
}

// IQSampleRate reads the sample rate for IQ data.
func (c *Client) IQSampleRate() (IQSampleRate, error) {
	reply, err := c.command("iq_samplerate")
	if err != nil {
		return 0, err
	}
	sampleRate, err := reply.ToInt(0)
	return IQSampleRate(sampleRate), err
}

// StartAudio starts the transmission of audio data for the given TRX.
func (c *Client) StartAudio(trx int) error {
	_, err := c.command("audio_start", trx)
	return err
}

// StopAudio stops the transmission of audio data for the given TRX.
func (c *Client) StopAudio(trx int) error {
	_, err := c.command("audio_stop", trx)
	return err
}

// SetAudioSampleRate sets sample rate for Audio data.
func (c *Client) SetAudioSampleRate(sampleRate AudioSampleRate) error {
	_, err := c.command("audio_samplerate", sampleRate)
	return err
}

// AudioSampleRate reads the sample rate for Audio data.
func (c *Client) AudioSampleRate() (AudioSampleRate, error) {
	reply, err := c.command("audio_samplerate")
	if err != nil {
		return 0, err
	}
	sampleRate, err := reply.ToInt(0)
	return AudioSampleRate(sampleRate), err
}

// AddSpot adds a spot to the panorama display.
func (c *Client) AddSpot(callsign string, mode Mode, frequency int, color ARGB, text string) error {
	_, err := c.command("spot", callsign, mode, frequency, color, text)
	return err
}

// DeleteSpot deletes the spot with the given callsign.
func (c *Client) DeleteSpot(callsign string) error {
	_, err := c.command("spot", callsign)
	return err
}

// ClearSpots deletes all spots.
func (c *Client) ClearSpots() error {
	_, err := c.command("spot_clear")
	return err
}

// SetVolume sets the main volume in dB.
func (c *Client) SetVolume(dB int) error {
	_, err := c.command("volume", dB)
	return err
}

// Volume reads the main volume in dB.
func (c *Client) Volume() (int, error) {
	reply, err := c.command("volume")
	if err != nil {
		return 0, err
	}
	return reply.ToInt(0)
}

// SetSquelchEnable enables the given TRX's squelch.
func (c *Client) SetSquelchEnable(trx int, enabled bool) error {
	_, err := c.command("sql_enable", trx, enabled)
	return err
}

// SquelchEnable reads the enable state of the given TRX's squelch.
func (c *Client) SquelchEnable(trx int) (bool, error) {
	reply, err := c.command("sql_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetSquelchLevel sets given TRX's squelch threshold in dB.
func (c *Client) SetSquelchLevel(dB int) error {
	_, err := c.command("sql_level", dB)
	return err
}

// SquelchLevel reads the given TRX's squelch threshold in dB.
func (c *Client) SquelchLevel() (int, error) {
	reply, err := c.command("sql_level")
	if err != nil {
		return 0, err
	}
	return reply.ToInt(0)
}

// SetVFOFrequency sets the tuning frequency of the given TRX's vfo.
func (c *Client) SetVFOFrequency(trx int, vfo VFO, frequency int) error {
	_, err := c.command("vfo", trx, vfo, frequency)
	return err
}

// VFOFrequency reads the tuning frequency of the given TRX's vfo.
func (c *Client) VFOFrequency(trx int, vfo VFO) (int, error) {
	reply, err := c.command("vfo", trx, vfo)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(2)
}

// BringToFront brings main ExpertSDR window into the focus.
func (c *Client) BringToFront() error {
	_, err := c.command("set_in_focus")
	return err
}

// SetMute mutes the main volume.
func (c *Client) SetMute(muted bool) error {
	_, err := c.command("mute", muted)
	return err
}

// Mute reads main volume's mute state.
func (c *Client) Mute() (bool, error) {
	reply, err := c.command("mute")
	if err != nil {
		return false, err
	}
	return reply.ToBool(0)
}

// SetRXMute mutes the given TRX's receiver.
func (c *Client) SetRXMute(trx int, muted bool) error {
	_, err := c.command("rx_mute", trx, muted)
	return err
}

// RXMute reads given TRX's receiver mute state.
func (c *Client) RXMute(trx int) (bool, error) {
	reply, err := c.command("rx_mute", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetCTCSSEnable enables CTCSS for the given TRX.
func (c *Client) SetCTCSSEnable(trx int, muted bool) error {
	_, err := c.command("ctcss_enable", trx, muted)
	return err
}

// CTCSSEnable reads enable state of CTCSS for the given TRX.
func (c *Client) CTCSSEnable(trx int) (bool, error) {
	reply, err := c.command("ctcss_enable", trx)
	if err != nil {
		return false, err
	}
	return reply.ToBool(1)
}

// SetCTCSSMode sets the CTCSS mode of the given TRX.
func (c *Client) SetCTCSSMode(trx int, mode CTCSSMode) error {
	_, err := c.command("ctcss_mode", trx, mode)
	return err
}

// CTCSSMode reads the CTCSS mode of the given TRX.
func (c *Client) CTCSSMode(trx int) (CTCSSMode, error) {
	reply, err := c.command("ctcss_enable", trx)
	if err != nil {
		return 0, err
	}
	mode, err := reply.ToInt(1)
	return CTCSSMode(mode), err
}

// SetCTCSSRXTone sets the given TRX's CTCSS subtone for receiving.
func (c *Client) SetCTCSSRXTone(trx int, tone CTCSSTone) error {
	_, err := c.command("ctcss_rx_tone", trx, tone)
	return err
}

// CTCSSRXTone reads the given TRX's CTCSS subtone for receiving.
func (c *Client) CTCSSRXTone(trx int) (CTCSSTone, error) {
	reply, err := c.command("ctcss_rx_tone", trx)
	if err != nil {
		return 0, err
	}
	tone, err := reply.ToInt(1)
	return CTCSSTone(tone), err
}

// SetCTCSSTXTone sets the given TRX's CTCSS subtone for transmitting.
func (c *Client) SetCTCSSTXTone(trx int, tone CTCSSTone) error {
	_, err := c.command("ctcss_tx_tone", trx, tone)
	return err
}

// CTCSSTXTone reads the given TRX's CTCSS subtone for transmitting.
func (c *Client) CTCSSTXTone(trx int) (CTCSSTone, error) {
	reply, err := c.command("ctcss_tx_tone", trx)
	if err != nil {
		return 0, err
	}
	tone, err := reply.ToInt(1)
	return CTCSSTone(tone), err
}

// SetCTCSSLevel sets the given TRX's CTCSS subtone level for transmitting in percent.
func (c *Client) SetCTCSSLevel(trx int, percent int) error {
	_, err := c.command("ctcss_level", trx, percent)
	return err
}

// CTCSSLevel reads the given TRX's CTCSS subtone level for transmitting in percent.
func (c *Client) CTCSSLevel(trx int) (int, error) {
	reply, err := c.command("ctcss_level", trx)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetECoderSwitchRX assigns the given TRX's control to the given E-Coder.
func (c *Client) SetECoderSwitchRX(ecoder int, trx int) error {
	_, err := c.command("ecoder_switch_rx", ecoder, trx)
	return err
}

// ECoderSwitchRX reads which TRX is assigned to the given E-Coder.
func (c *Client) ECoderSwitchRX(ecoder int) (int, error) {
	reply, err := c.command("ecoder_switch_rx", ecoder)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(1)
}

// SetECoderSwitchChannel assigns the given channel's control to the given E-Coder.
func (c *Client) SetECoderSwitchChannel(ecoder int, vfo VFO) error {
	_, err := c.command("ecoder_switch_channel", ecoder, vfo)
	return err
}

// ECoderSwitchChannel reads which channel is assigned to the given E-Coder.
func (c *Client) ECoderSwitchChannel(ecoder int) (VFO, error) {
	reply, err := c.command("ecoder_switch_channel", ecoder)
	if err != nil {
		return 0, err
	}
	vfo, err := reply.ToInt(1)
	return VFO(vfo), err
}

// SetRXVolume sets the given TRX's channel volume in dB.
func (c *Client) SetRXVolume(trx int, vfo VFO, dB int) error {
	_, err := c.command("rx_volume", trx, vfo, dB)
	return err
}

// RXVolume reads the given TRX's channel volume in dB.
func (c *Client) RXVolume(trx int, vfo VFO) (int, error) {
	reply, err := c.command("rx_volume", trx, vfo)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(2)
}

// SetRXBalance sets the given TRX's channel balance in dB.
func (c *Client) SetRXBalance(trx int, vfo VFO, dB int) error {
	_, err := c.command("rx_balance", trx, vfo, dB)
	return err
}

// RXBalance reads the given TRX's channel balance in dB.
func (c *Client) RXBalance(trx int, vfo VFO) (int, error) {
	reply, err := c.command("rx_balance", trx, vfo)
	if err != nil {
		return 0, err
	}
	return reply.ToInt(2)
}
