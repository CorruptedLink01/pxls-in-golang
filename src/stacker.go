package main

import (
	"context"
	"time"
)

// PixelStacker increases the user's available pixels over time.
// It uses the channel C to communicate that the Stack has changed:
// - if it sends `true`, the Stack gained a pixel
// - if it sends `false`, the Stack was consumed
type PixelStacker struct {
	ctx         context.Context
	ctxCancel   context.CancelFunc
	CooldownEnd time.Time
	Stack       uint
	C           chan bool
}

func (ps *PixelStacker) getAndUpdateCooldown() (cd time.Duration) {
	cd = ps.GetCooldown()
	ps.CooldownEnd = time.Now().Add(cd)
	return
}

func (ps *PixelStacker) run() {
	var max = uint(App.Conf.GetInt32("stacking.maxStacked"))
	cd := ps.getAndUpdateCooldown()

	// Note(netux): <= instead of < is intentional
	for ps.Stack <= max {
		select {
		case <-time.After(cd):
			ps.Gain()
			cd = ps.getAndUpdateCooldown()
		case <-ps.ctx.Done():
			return
		}
	}
}

// GetCooldown returns the user's cooldown in between receiving
// available pixels based on how many pixels they've got
// available already, and a multiplicative factor.
func (ps *PixelStacker) GetCooldown() time.Duration {
	// TODO(netux): check if the second stacked pixel has twice the factor
	var factor = float32(App.Conf.GetFloat32("stacking.cooldownMultiplier"))
	return time.Duration(float32(ps.Stack+1)*factor) * App.GetCooldown()
}

// GetCooldownWithDifference returns the user's cooldown that is left
// since the last pixel gain.
func (ps *PixelStacker) GetCooldownWithDifference() (cd time.Duration) {
	var now = time.Now()
	cd = ps.GetCooldown()
	if now.Before(ps.CooldownEnd) {
		cd -= cd - ps.CooldownEnd.Sub(now)
	}
	return cd
}

// StartTimer starts the PixelStacker
func (ps *PixelStacker) StartTimer() {
	if ps.IsTimerRunning() {
		ps.ctxCancel()
	}
	ps.ctx, ps.ctxCancel = context.WithCancel(context.Background())
	go ps.run()
}

// StopTimer stops the PixelStacker
func (ps *PixelStacker) StopTimer() {
	ps.ctxCancel()
}

// IsTimerRunning returns whenever the pixel stacker's timer is running
func (ps *PixelStacker) IsTimerRunning() bool {
	return ps.ctx != nil && ps.ctx.Err() != context.Canceled
}

// Gain increases the stack and notifies that through the channel C.
func (ps *PixelStacker) Gain() {
	if ps.Stack <= uint(App.Conf.GetInt32("stacking.maxStacked")) {
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
	ps := PixelStacker{
		C: make(chan bool, 1),
	}

	return &ps
}
