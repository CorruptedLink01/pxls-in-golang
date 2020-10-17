package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/go-akka/configuration"
)

const (
	// CanvasBoardFile is the name of the canvas board file
	CanvasBoardFile = "board.dat"

	// MaxUserAmount is the maximum amount of concurrent users supported by the server
	MaxUserAmount = 512

	// TODO(netux): lower these values

	// MaxWebsocketReadBufferSize is the maximum size limit of a valid websocket incoming message
	MaxWebsocketReadBufferSize = 1024
	// MaxWebsocketSendBufferSize is the maximum size limit of a valid websocket outgoing message
	MaxWebsocketSendBufferSize = 1024
)

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

// makeCanvasFromConf reads width and height from the config file and creates a canvas
func makeCanvasFromConf(conf *configuration.Config) (*Canvas, error) {
	return NewCanvas(
		uint(conf.GetInt32("board.width")),
		uint(conf.GetInt32("board.height")),
		byte(conf.GetInt32("board.defaultColor")),
	), nil
}

// populateCanvasFromFile reads the canvas board file and writes
// its contents to the canvas board.
func populateCanvasFromFile(c *Canvas) error {
	b, err := ioutil.ReadFile(CanvasBoardFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s not found, using blank board\n", CanvasBoardFile)
			return nil
		}
		return err
	}

	if uint(len(b)) > c.Width*c.Height {
		// TODO(netux): implement input handling to opt-out of this
		fmt.Printf("saved %s size and canvas configuration differ, which means canvas might be corrupted\n", CanvasBoardFile)
		fmt.Printf("using %s up to the configured size in 20 seconds unless this program is terminated\n", CanvasBoardFile)

		<-time.After(20 * time.Second)
	}

	c.Board = b[:c.Width*c.Height]
	return nil
}

// saveCanvas writes the contents of the board
// into the canvas board file
func saveCanvas(c *Canvas) error {
	var mode = os.FileMode(644)

	fi, err := os.Stat(CanvasBoardFile)
	if err == nil {
		mode = fi.Mode()
	} else if !os.IsNotExist(err) {
		return err
	}

	err = ioutil.WriteFile(CanvasBoardFile, c.Board, mode)
	return err
}

// saveCanvasEvery calls saveCanvas every d time.
func saveCanvasEvery(c *Canvas, d time.Duration) {
	err := saveCanvas(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "save canvas board err: %v", err)
	}

	<-time.After(d)

	saveCanvasEvery(c, d)
}

// App stores globally accesible information about the game application
var App PxlsApp

func main() {
	conf, err := ReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading config err: %v\n", err)
		return
	}

	db, err := MakeDatabase(
		conf.GetString("database.driver"),
		conf.GetString("database.user"),
		conf.GetString("database.pass"),
		conf.GetString("database.url"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initializing database err: %v\n", err)
		return
	}
	defer db.Close()

	palette, err := makePaletteFromConf(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "palette parsing from config err: %v\n", err)
		return
	}

	canvas, err := makeCanvasFromConf(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "canvas parsing from config err: %v\n", err)
		return
	}
	populateCanvasFromFile(canvas)

	App = PxlsApp{*conf, *db, *canvas, palette, *MakeUserList()}

	go saveCanvasEvery(canvas, conf.GetTimeDurationInfiniteNotAllowed("board.saveInterval", 5*time.Second))

	go StartCommands(canvas)

	StartServer()
}
