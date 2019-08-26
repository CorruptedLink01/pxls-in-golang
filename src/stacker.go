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
	ctx       context.Context
	ctxCancel context.CancelFunc
	Stack     uint
	C         chan bool
}

func (ps *PixelStacker) run() {
	var max = uint(App.Conf.GetInt32("stacking.maxStacked"))

	// Note(netux): <= instead of < is intentional
	for ps.Stack <= max {
		select {
		case <-time.After(ps.GetCooldown()):
			ps.Gain()
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

// StartTimer starts the PixelStacker
func (ps *PixelStacker) StartTimer() {
	if ps.ctx != nil && ps.ctx.Err() != context.Canceled {
		ps.ctxCancel()
	}
	ps.ctx, ps.ctxCancel = context.WithCancel(context.Background())
	go ps.run()
}

// StopTimer stops the PixelStacker
func (ps *PixelStacker) StopTimer() {
	ps.ctxCancel()
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
