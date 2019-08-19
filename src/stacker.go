package main

import (
	"time"
)

// PixelStacker increases the user's available pixels over time.
// It uses the channel C to communicate that the Stack has changed:
// - if it sends `true`, the Stack gained a pixel
// - if it sends `false`, the Stack was consumed
type PixelStacker struct {
	timer *time.Timer
	stop  chan bool
	Stack uint
	C     chan bool
}

func (ps *PixelStacker) run() {
	var max = uint(App.conf.GetInt32("stacking.maxStacked"))

	// TODO(netux): check if calling ps.timer.Reset() is more efficient
	ps.timer = time.NewTimer(ps.GetCooldown())

	select {
	case <-ps.timer.C:
		// Note(netux): > instead of == is intentional
		if ps.Stack > max {
			return
		}

		ps.Gain()
		ps.run()
	case <-ps.stop:
		return
	}
}

// GetCooldown returns the user's cooldown in between receiving
// available pixels based on how many pixels they've got
// available already, and a multiplicative factor.
func (ps *PixelStacker) GetCooldown() time.Duration {
	// TODO(netux): check if the second stacked pixel has twice the factor
	var factor = float32(App.conf.GetFloat32("stacking.cooldownMultiplier"))
	return time.Duration(float32(ps.Stack+1)*factor) * App.GetCooldown()
}

// StartTimer starts the PixelStacker
func (ps *PixelStacker) StartTimer() {
	// TODO(netux): this feels hacky
	if ps.timer != nil {
		ps.timer.Stop()
		go func() {
			for <-ps.stop {
				// drain ps.stop
			}
		}()
	}
	go ps.run()
}

// StopTimer stops the PixelStacker
func (ps *PixelStacker) StopTimer() {
	go func() {
		ps.stop <- true
	}()
}

// Gain increases the stack and notifies that through the channel C.
func (ps *PixelStacker) Gain() {
	if ps.Stack <= uint(App.conf.GetInt32("stacking.maxStacked")) {
		ps.Stack++
		ps.C <- true
	}
}

// Consume decreases the stack and notifies that through the channel C.
func (ps *PixelStacker) Consume() {
	if ps.Stack > 0 {
		ps.Stack--
		ps.C <- false
	}
}

// MakePixelStacker creates a new, clean, PixelStacker
func MakePixelStacker() *PixelStacker {
	return &PixelStacker{
		timer: nil,
		stop:  make(chan bool, 1),
		C:     make(chan bool, 1),
	}
}
