package cmd

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/spf13/cobra"

	"github.com/ftl/tci/client"
)

var toneFlags = struct {
	frequency1 float64
	amplitude1 float64
	phase1     float64
	frequency2 float64
	amplitude2 float64
	phase2     float64
	timeout    time.Duration
}{}

var toneCmd = &cobra.Command{
	Use:   "tone",
	Short: "A simple tone generator with one or two tones.",
	Run:   runWithClient(tone),
}

func init() {
	rootCmd.AddCommand(toneCmd)

	toneCmd.Flags().Float64Var(&toneFlags.frequency1, "f1", 1000, "the primary frequency")
	toneCmd.Flags().Float64Var(&toneFlags.amplitude1, "a1", 1, "the primary amplitude (0-1)")
	toneCmd.Flags().Float64Var(&toneFlags.phase1, "p1", 0, "the primary phase (p1 * Pi)")
	toneCmd.Flags().Float64Var(&toneFlags.frequency2, "f2", 0, "the secondary frequency (zero for single tone)")
	toneCmd.Flags().Float64Var(&toneFlags.amplitude2, "a2", 1, "the secondary amplitude (0-1)")
	toneCmd.Flags().Float64Var(&toneFlags.phase2, "p2", 0, "the secondary phase (p2 * Pi)")
	toneCmd.Flags().DurationVar(&toneFlags.timeout, "timeout", 0, "the timeout")
}

func tone(ctx context.Context, c *client.Client, _ *cobra.Command, _ []string) {
	sampleRate, err := c.AudioSampleRate()
	if err != nil {
		log.Fatalf("cannot get audio sample rate: %v", err)
	}

	osc := newOscillator(c, int(sampleRate))
	log.Printf("tone generator f1=%f, a1=%f, p1=%f, f2=%f, a2=%f, p2=%f",
		osc.frequency1,
		osc.amplitude1,
		osc.phase1,
		osc.frequency2,
		osc.amplitude2,
		osc.phase2,
	)
	c.Notify(osc)

	c.StartAudio(0)
	defer c.StopAudio(0)
	c.SetTX(0, true, client.SignalSourceVAC)
	defer c.SetTX(0, false, client.SignalSourceVAC)

	if toneFlags.timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, toneFlags.timeout)
		defer cancel()
	}
	<-ctx.Done()
	c.SetTX(0, false, client.SignalSourceVAC)
}

type oscillator struct {
	c    *client.Client
	buf  []float32
	tick float64
	t    float64

	frequency1 float64
	amplitude1 float64
	phase1     float64
	frequency2 float64
	amplitude2 float64
	phase2     float64
}

func newOscillator(c *client.Client, sampleRate int) *oscillator {
	return &oscillator{
		c: c,

		frequency1: toneFlags.frequency1,
		amplitude1: toneFlags.amplitude1,
		phase1:     toneFlags.phase1 * math.Pi, // 0 = 0.0*math.Pi; 90 = 0.5*math.Pi; 180 = 1.0*math.Pi; 270 = 1.5*math.Pi;

		frequency2: toneFlags.frequency2,
		amplitude2: toneFlags.amplitude2,
		phase2:     toneFlags.phase2 * math.Pi, // 0 = 0.0*math.Pi; 90 = 0.5*math.Pi; 180 = 1.0*math.Pi; 270 = 1.5*math.Pi;

		tick: 1.0 / float64(sampleRate),
	}
}

func (o *oscillator) Read(out []float32) (int, error) {
	for i := 0; i < len(out)-1; i += 2 {
		out[i] = float32(o.amplitude1 * math.Cos(2*math.Pi*o.frequency1*o.t+o.phase1))
		if o.frequency2 != 0 {
			out[i] += float32(o.amplitude2 * math.Cos(2*math.Pi*o.frequency2*o.t+o.phase2))
		}
		out[i+1] = out[i]
		o.t += o.tick
	}
	return len(out), nil
}

func (o *oscillator) TXChrono(trx int, sampleRate client.AudioSampleRate, requestedSampleCount uint32) {
	if len(o.buf) != int(requestedSampleCount) {
		o.buf = make([]float32, requestedSampleCount)
	}
	sampleCount, err := o.Read(o.buf)
	if err != nil {
		log.Printf("cannot generate tx audio: %v", err)
	}
	err = o.c.SendTXAudio(trx, sampleRate, o.buf[0:sampleCount])
	if err != nil {
		log.Printf("cannot send tx audio: %v", err)
	}
}
