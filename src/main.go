package main

import (
	"fmt"
	"strconv"

	"github.com/go-akka/configuration"
)

const (
	// MaxUserAmount is the maximum amount of concurrent users supported by the server
	MaxUserAmount = 512

	// TODO(netux): lower these values

	// MaxWebsocketReadBufferSize is the maximum size limit of a valid websocket incoming message
	MaxWebsocketReadBufferSize = 1024
	// MaxWebsocketSendBufferSize is the maximum size limit of a valid websocket outgoing message
	MaxWebsocketSendBufferSize = 1024
)

// makeCanvasFromConf reads width and height from the config file and creates a canvas
func makeCanvasFromConf(conf *configuration.Config) (*Canvas, error) {
	return NewCanvas(
		uint(conf.GetInt32("board.width")),
		uint(conf.GetInt32("board.height")),
		byte(conf.GetInt32("board.defaultColor")),
	), nil
}

// makePaletteFromConf converts the palette in the config file to a Palette
func makePaletteFromConf(conf *configuration.Config) (Palette, error) {
	cPalette := conf.GetStringList("board.palette")

	var palette = make(Palette, len(cPalette))
	for i, v := range cPalette {
		c, err := strconv.ParseInt(v[1:], 16, 32)
		if err != nil {
			return nil, err
		}
		palette[i] = int(c)
	}

	return palette, nil
}

// App stores globally accesible information about the game application
var App PxlsApp

func main() {
	conf, err := ReadConfig()
	if err != nil {
		fmt.Printf("Reading config err: %s\n", err)
		return
	}

	palette, err := makePaletteFromConf(conf)
	if err != nil {
		fmt.Printf("Palette parsing from config err: %s\n", err)
		return
	}
	canvas, err := makeCanvasFromConf(conf)
	if err != nil {
		fmt.Printf("Canvas parsing from config err: %s\n", err)
		return
	}

	var users = make([]User, 0, MaxUserAmount)

	App = PxlsApp{*conf, *canvas, palette, users}

	StartServer()
}
