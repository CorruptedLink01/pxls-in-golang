package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiAuthServices struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type apiInfo struct {
	CanvasCode          string                     `json:"canvasCode"`
	Width               uint                       `json:"width"`
	Height              uint                       `json:"height"`
	Palette             []string                   `json:"palette"`
	CaptchaKey          string                     `json:"captchaKey"`
	HeatmapCooldown     int                        `json:"heatmapCooldown"`
	MaxStackedPixels    uint                       `json:"maxStacked"`
	AuthServices        map[string]apiAuthServices `json:"authServices"`
	RegistrationEnabled bool                       `json:"registrationEnabled"`
}

type apiWhoAmI struct {
	Name string `json:"string"`
	ID   int    `json:"id"`
}

func intToHex(cs []int) []string {
	res := make([]string, len(cs))
	for i, c := range cs {
		res[i] = fmt.Sprintf("#%06X", c)
	}
	return res
}

// TODO(netux): replace fmt with a logger

// StartServer sets up endpoint handlers and listens and serves
func StartServer() {
	// handle /info
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := apiInfo{
			CanvasCode:       App.Conf.GetString("canvascode"),
			Width:            App.Canvas.Width,
			Height:           App.Canvas.Height,
			Palette:          intToHex(App.Palette),
			MaxStackedPixels: uint(App.Conf.GetInt32("stacking.maxStacked")),
			// TODO(netux): return actually supported auth services
			AuthServices: map[string]apiAuthServices{
				"discord": apiAuthServices{
					ID:   "discord",
					Name: "Discord",
				},
			},
			RegistrationEnabled: false,
		}

		w.Header().Set("Content-Type", "text/json")
		json.NewEncoder(w).Encode(info)
	})

	// handle /boarddata
	http.HandleFunc("/boarddata", func(w http.ResponseWriter, r *http.Request) {
		w.Write(App.Canvas.Board)
	})

	// handle /whoami
	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		var res = apiWhoAmI{"-snip-", -1}
		json.NewEncoder(w).Encode(res)
	})

	// handle /ws
	http.HandleFunc("/ws", HandleWebsocketPath)

	// handle static
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	port := App.Conf.GetString("server.port")
	http.ListenAndServe(":"+port, nil)
}
