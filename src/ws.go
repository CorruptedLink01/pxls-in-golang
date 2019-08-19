package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// TODO(netux): implement something similar to this
// https://github.com/gorilla/websocket/tree/master/examples/chat

type wsConn struct {
	*websocket.Conn
	pxlsUser  *User
	sendQueue chan interface{}
}

func (conn *wsConn) queue(msg interface{}) {
	conn.sendQueue <- msg
}

type wsMessageType string
type wsMessage struct {
	Type wsMessageType `json:"type"`
}

func withType(t wsMessageType) wsMessage {
	return wsMessage{t}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  MaxWebsocketReadBufferSize,
	WriteBufferSize: MaxWebsocketSendBufferSize,
}

// Connections is an array of all active websocket connections.
var Connections = make(map[string]*wsConn)

func upgradeSocket(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	var confAuthUseIP = App.conf.GetBoolean("oauth.useIp")
	pxlsTokenCookie, err := r.Cookie("pxls-token")

	var user *User
	for _, u := range App.users {
		// match IP if connected confAuthUseIP or match token otherwise
		var match bool
		if confAuthUseIP {
			host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("User sent invalid address %s", conn.RemoteAddr())
			}
			match = u.Auth.IP == host
		} else if pxlsTokenCookie != nil {
			match = u.Auth.Token == pxlsTokenCookie.Value
		}

		if match {
			user = &u
			break
		}
	}
	if user == nil && confAuthUseIP {
		user = MakeUserFromIP(uint(len(App.users)), conn.LocalAddr().String())
		App.users = append(App.users, *user)
	}

	return &wsConn{
		conn,
		user,
		make(chan interface{}),
	}, nil
}

// HandleWebsocketPath is an http.HandleFunc which upgrades the request,
// setups the connection and starts handling websocket messages
func HandleWebsocketPath(w http.ResponseWriter, r *http.Request) {
	conn, err := upgradeSocket(w, r)
	if err != nil {
		fmt.Printf("Websocket upgrade err: %s\n", err)
		return
	}

	Connections[conn.RemoteAddr().String()] = conn
	go func() {
		for {
			conn.WriteJSON(<-conn.sendQueue)
		}
	}()

	if conn.pxlsUser != nil {
		// Note(netux): Needed so the max stacked on the client updates
		sendPixelsAvailable(conn, "auth")
		sendUserInfo(conn)
		if conn.pxlsUser.PixelStacker.Stack > 0 {
			sendPixelsAvailable(conn, "connected")
		}
		conn.pxlsUser.PixelStacker.StartTimer()
		if conn.pxlsUser.PixelStacker.Stack == 0 {
			sendCooldown(conn, conn.pxlsUser.PixelStacker.GetCooldown())
		}
	}

	go handleIncomingMessages(conn)
	if conn.pxlsUser != nil {
		go handleUserEvents(conn)
	}
}

func handleUserEvents(conn *wsConn) {
	for {
		isGain := <-conn.pxlsUser.PixelStacker.C
		cause := "consume"
		if isGain {
			cause = "stackGain"
		}
		sendPixelsAvailable(conn, cause)
	}
}

func handleIncomingMessages(conn *wsConn) {
	defer func() {
		delete(Connections, conn.RemoteAddr().String())
		conn.Close()
	}()

	for {
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				return
			}
			fmt.Printf("Conn reading err: %s\n", err)
			continue
		}
		var msgType wsMessageType
		{
			var wsMsg wsMessage
			if err := json.Unmarshal(rawMsg, &wsMsg); err != nil {
				fmt.Printf("Websocket JSON parsing error: %s\n", err)
				return
			}
			msgType = wsMsg.Type
		}

		switch msgType {
		case wsPixelType:
			if conn.pxlsUser == nil {
				break
			}

			var pixelMsg wsPixelReq
			if err := json.Unmarshal(rawMsg, &pixelMsg); err != nil {
				fmt.Printf("Websocket JSON Pixel parsing error: %s\n", err)
				break
			}
			handlePixel(conn, pixelMsg)
		default:
			fmt.Printf("Unhandled websocket msgType %s: %s\n", msgType, string(rawMsg))
		}
	}
}

const wsUserInfoType wsMessageType = "userinfo"

type wsUserInfo struct {
	wsMessage
	AuthMethod string   `json:"method"`
	Role       UserRole `json:"role"`
	Username   string   `json:"username"`

	CooldownOverride bool `json:"cdOverride"`

	IsBanned           bool   `json:"banned"`
	BanExpiry          int    `json:"banExpiry"`
	BanReason          string `json:"ban_reason"`
	IsChatBanned       bool   `json:"chatBanned"`
	ChatBanExpiry      int    `json:"chatBanExpiry"`
	ChatBanIsPermanent bool   `json:"chatBanIsPerma"`
}

// sendUserInfo sends an "userinfo" message through the websocket connection conn
func sendUserInfo(conn *wsConn) {
	// TODO(netux): support non-ip method
	conn.queue(wsUserInfo{
		wsMessage:  withType(wsUserInfoType),
		AuthMethod: conn.pxlsUser.Auth.Method,
		Role:       conn.pxlsUser.Role,
		Username:   conn.pxlsUser.Name,
	})
}

const wsAckType = "ACK"

type wsAck struct {
	wsMessage
	ackFor string
}

func ackFor(a string) wsAck {
	return wsAck{withType(wsAckType), a}
}

const wsPixelsAvailableType = "pixels"

type wsPixelsAvailable struct {
	wsMessage
	Count uint   `json:"count"`
	Cause string `json:"cause"`
}

func sendPixelsAvailable(conn *wsConn, cause string) {
	conn.queue(wsPixelsAvailable{
		withType(wsPixelsAvailableType),
		conn.pxlsUser.PixelStacker.Stack,
		cause,
	})
}

const wsCooldownType = "cooldown"

type wsCooldown struct {
	wsMessage
	Wait float32 `json:"wait"`
}

func sendCooldown(conn *wsConn, wait time.Duration) {
	conn.queue(wsCooldown{
		withType(wsCooldownType),
		float32(wait) / float32(time.Second),
	})
}

const wsPixelType = "pixel"

type wsPixel struct {
	PosX     uint `json:"x"`
	PosY     uint `json:"y"`
	ColorIdx byte `json:"color"`
}

type wsPixelReq struct {
	wsMessage
	wsPixel
}

type wsPixelRes struct {
	wsMessage
	Pixels []wsPixel `json:"pixels"`
}

type wsAckForPixel struct {
	wsAck
	PosX uint `json:"x"`
	PosY uint `json:"y"`
}

func handlePixel(conn *wsConn, pixelMsg wsPixelReq) {
	if conn.pxlsUser == nil {
		return
	}

	var ps = conn.pxlsUser.PixelStacker
	if ps.Stack == 0 {
		return
	}

	if App.canvas.GetPixelColorIndex(pixelMsg.PosX, pixelMsg.PosY) == pixelMsg.ColorIdx {
		return
	}

	ps.StopTimer()
	conn.queue(wsAckForPixel{
		ackFor("PLACE"),
		pixelMsg.PosX,
		pixelMsg.PosY,
	})

	ps.Consume()

	App.canvas.SetPixelColor(pixelMsg.PosX, pixelMsg.PosY, pixelMsg.ColorIdx)
	ps.StartTimer()

	if conn.pxlsUser.PixelStacker.Stack == 0 {
		sendCooldown(conn, ps.GetCooldown())
	}

	pixelsMsg := wsPixelRes{
		wsMessage{
			Type: wsPixelType,
		},
		[]wsPixel{pixelMsg.wsPixel},
	}
	for _, conn := range Connections {
		conn.queue(pixelsMsg)
	}
}
