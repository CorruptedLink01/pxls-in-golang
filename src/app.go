package main

import (
	"time"

	"github.com/go-akka/configuration"
)

// Palette is a list of color values that can be used in the canvas.
type Palette []int

// Canvas contains canvas information: width, height and pixel color indices.
type Canvas struct {
	Width  uint
	Height uint
	Board  []byte
}

// NewCanvas creates new canvas with the width and height specified.
func NewCanvas(w, h uint, bgColorIdx byte) *Canvas {
	c := &Canvas{w, h, make([]byte, w*h)}
	for y := uint(0); y < h; y++ {
		for x := uint(0); x < w; x++ {
			c.Board[x+y*w] = bgColorIdx
		}
	}
	return c
}

// GetPixelColorIndex returns the color index of the pixel.
func (c *Canvas) GetPixelColorIndex(x, y uint) byte {
	return c.Board[x+y*c.Width]
}

// SetPixelColor sets the color index of a pixel.
func (c *Canvas) SetPixelColor(x, y uint, colorIdx byte) {
	c.Board[x+y*c.Width] = colorIdx
}

// PxlsApp stores information about the game application.
type PxlsApp struct {
	Conf    configuration.Config
	DB      Database
	Canvas  Canvas
	Palette Palette
	Users   UserList
}

// GetCooldown returns the time in between placing pixels
// players have to wait until they can place again.
func (a *PxlsApp) GetCooldown() time.Duration {
	var cooldown = a.Conf.GetTimeDurationInfiniteNotAllowed("cooldown")
	// TODO(netux): apply math function used in Pxls' sourcecode
	return cooldown
}
